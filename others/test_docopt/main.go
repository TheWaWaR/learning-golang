package main

import (
	"fmt"
	"github.com/docopt/docopt-go"
)


var (
	USAGE = `Test docopt
Usage:
  {prog} generate [dates <date>... | range <date-start> <date-end>] -d=<log-dir> -c=<cache-dir>
      [-r=<auto-report> -l=<model-limit> -D]
  {prog} report [dates <date>... | range <date-start> <date-end>] -c=<cache-dir>
      [-d=<log-dir> -l=<model-limit> -D]

Options:
  -D --debug             Debug mode
  -d --log-dir DIR       Log files (<DIR>/YYYY/m/yyy-mm-dd.log)
  -r --auto-report BOOL  If automatic report the result [default: yes]
  -l --model-limit NUM   Model pie chart limit [default: 10]
  -c --cache-dir DIR     Cache directory to save result and mapping files
  -h --help              Show this screen
`
)


func main() {
	arguments, _ := docopt.Parse(USAGE, nil, true, "Naval Fate 2.0", false)
	fmt.Printf("%+v\n", arguments)
}
