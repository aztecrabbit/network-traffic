package main

import (
	"os"
	"fmt"
	"time"
	"strconv"
	"strings"
	"io/ioutil"
)

var (
	interface_names []string = os.Args[1:]
	interval int = 1
)

func GetSize(data int) string {
	suffixes := []string{
		"KB",
		"KB",
		"MB",
		"GB",
	}
	value := float64(data)
	i := 0

	for value >= 1024 && i < (len(suffixes) - 1) {
		value = value / 1024
		i++
	}

	if i == 0 {
		return fmt.Sprintf("0.%.3d %s", int(value), suffixes[i])
	}

	return fmt.Sprintf("%.3f %s", value, suffixes[i])
}

func GetNetworkStatistic(interface_name string, id string) int {
	data, err := ioutil.ReadFile(fmt.Sprintf("/sys/class/net/%s/statistics/%s_bytes", interface_name, id))
	if err != nil {
		return -1
	}

	bytes, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		panic(err)
	}

	return bytes
}

func main() {
	if len(interface_names) == 0 {
		fmt.Println("Usage: network-traffic [interface] [interface] [...]")
		os.Exit(0)
	}

	network := make(map[string]map[string]int)

	// rx = down
	// tx = up

	for _, interface_name := range interface_names {
		network[interface_name] = make(map[string]int)
		network[interface_name]["rx_pas"] = GetNetworkStatistic(interface_name, "rx")
		network[interface_name]["tx_pas"] = GetNetworkStatistic(interface_name, "tx")
	}


	for {
		time.Sleep(time.Duration(interval) * time.Second)
		data := make([]string, 0)

		for _, interface_name := range interface_names {
			network[interface_name]["rx_now"] = GetNetworkStatistic(interface_name, "rx")
			network[interface_name]["tx_now"] = GetNetworkStatistic(interface_name, "tx")

			if network[interface_name]["rx_now"] == -1 || network[interface_name]["tx_now"] == -1 {
				continue
			}

			network[interface_name]["rx_statistic"] = (
				(network[interface_name]["rx_now"] - network[interface_name]["rx_pas"]) / interval)
			network[interface_name]["tx_statistic"] = (
				(network[interface_name]["tx_now"] - network[interface_name]["tx_pas"]) / interval)
			network[interface_name]["rx_pas"] = network[interface_name]["rx_now"]
			network[interface_name]["tx_pas"] = network[interface_name]["tx_now"]
			network[interface_name]["io_statistic_total"] = (
				network[interface_name]["rx_now"] + network[interface_name]["tx_now"])

			data = append(data, fmt.Sprintf("%s/s  %s/s  %s",
				GetSize(network[interface_name]["rx_statistic"]),
				GetSize(network[interface_name]["tx_statistic"]),
				GetSize(network[interface_name]["io_statistic_total"]),
			))
		}

		data = append(data, GetSize(
			GetNetworkStatistic("lo", "rx") + GetNetworkStatistic("lo", "tx"),
		))

		fmt.Println(strings.Join(data, "  -  "))
	}
}
