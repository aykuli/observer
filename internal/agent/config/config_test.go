package config

import (
	"fmt"
)

func Example() {
	address := ":5050"
	reportInterval := "66"
	pollInterval := "77"
	key := "aynur_is_beautiful_name_fit_to_be_key"
	rateLimit := "88"
	testFlags := map[string]string{
		"a": address,
		"r": reportInterval,
		"p": pollInterval,
		"k": key,
		"l": rateLimit}
	arr := make([]string, 6)
	i := 0
	for k, v := range testFlags {
		arr[i] = "-" + k + "=" + v
		i++
	}

	parseFlags(arr)

	fmt.Println(Options.Address)
	fmt.Println(Options.ReportInterval)
	fmt.Println(Options.PollInterval)
	fmt.Println(Options.Key)
	fmt.Println(Options.RateLimit)

	//	Output:
	// :5050
	// 66
	// 77
	// aynur_is_beautiful_name_fit_to_be_key
	// 88
}
