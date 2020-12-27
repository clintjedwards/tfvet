package cli

import (
	"log"
	"os"

	"github.com/clintjedwards/tfvet/cli/appcfg"
	"github.com/clintjedwards/tfvet/utils"
)

func runInit() error {

	err := utils.CreateDir(appcfg.ConfigPath())
	if err != nil {
		log.Println(err)
		return err
	}
	err = utils.CreateDir(appcfg.RulesetsPath())
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = os.Stat(appcfg.ConfigFilePath())
	if os.IsNotExist(err) {
		err = appcfg.CreateNewFile()
		if err != nil {
			log.Println(err)
			return err
		}
	} else if os.IsExist(err) {
	} else if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
