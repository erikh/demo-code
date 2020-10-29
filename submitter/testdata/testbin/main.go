package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"time"
)

func main() {
	sleep := flag.Int("sleep", 0, "how long to sleep (seconds)")
	hang := flag.Bool("hang", false, "whether to hang indefinitely or not")
	flag.Parse()

	in := map[string]interface{}{}

	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		panic(err)
	}

	if *hang {
		for {
			time.Sleep(time.Second)
		}
	}

	if *sleep != 0 {
		time.Sleep(time.Second * time.Duration(*sleep))
	}

	out := map[string]interface{}{}

	if len(flag.Args()) == 2 {
		content, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			panic(err)
		}

		if err := json.Unmarshal(content, &out); err != nil {
			panic(err)
		}
	}

	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		panic(err)
	}
}
