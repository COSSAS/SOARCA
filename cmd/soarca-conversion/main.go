package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"soarca/internal/logger"
	"soarca/pkg/conversion"
)

const banner = `
   _____  ____          _____   _____          
  / ____|/ __ \   /\   |  __ \ / ____|   /\    
 | (___ | |  | | /  \  | |__) | |       /  \   
  \___ \| |  | |/ /\ \ |  _  /| |      / /\ \  
  ____) | |__| / ____ \| | \ \| |____ / ____ \ 
 |_____/ \____/_/    \_\_|  \_\\_____/_/    \_\
                                               
                                               

`

var log *logger.Log

var (
	Version   string
	Buildtime string
	Host      string
)

func init() {
	log = logger.Logger("CONVERTER", logger.Info, "", logger.Json)
}
func main() {
	fmt.Print(banner)
	log.Info("Version: ", Version)
	log.Info("Buildtime: ", Buildtime)
	var (
		source_filename string
		target_filename string
		format          string
	)
	flag.StringVar(&source_filename, "source", "", "The source file to be converted")
	flag.StringVar(&target_filename, "target", "", "The name of the converted filename")
	flag.StringVar(&format, "format", "", "The format of the source file")
	flag.Parse()
	if source_filename == "" {
		log.Error("No source file given: -source=SOURCE_FILE is required")
		return
	}
	if target_filename == "" {
		target_filename = fmt.Sprintf("%s.json", source_filename)
		log.Infof("No target file given: defaulting to %s", target_filename)
	}
	source_content, err := os.ReadFile(source_filename)
	if err != nil {
		log.Errorf("Could not read source file")
	}
	target_content, err := conversion.PerformConversion(source_filename, source_content, format)
	if err != nil {
		log.Error(err)
	}
	output_str, err := json.Marshal(target_content)
	if err != nil {
		log.Error(err)
	}
	os.WriteFile(target_filename, output_str, 0644)

}
