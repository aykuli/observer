package config

import (
	"fmt"
)

func Example() {
	address := ":5050"
	fileStoragePath := "awesomePath.txt"
	storeInterval := "1"
	restore := "false"
	databaseDsn := "some_dsn_path"
	key := "aynur_is_beautiful_name_fit_to_be_key"
	testFlags := map[string]string{
		"a": address,
		"f": fileStoragePath,
		"i": storeInterval,
		"r": restore,
		"d": databaseDsn,
		"k": key}
	arr := make([]string, 6)
	i := 0
	for k, v := range testFlags {
		arr[i] = "-" + k + "=" + v
		i++
	}

	parseFlags(arr)

	fmt.Println(Options.Address)
	fmt.Println(Options.FileStoragePath)
	fmt.Println(Options.StoreInterval)
	fmt.Println(Options.Restore)
	fmt.Println(Options.DatabaseDsn)
	fmt.Println(Options.Key)

	//	Output:
	// :5050
	// awesomePath.txt
	// 1
	// false
	// some_dsn_path
	// aynur_is_beautiful_name_fit_to_be_key
}
