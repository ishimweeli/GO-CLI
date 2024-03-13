package dirconfig

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"amalitech.org/subsys/utils"
)

type Configurator struct {
	Interactive bool
	Override    bool
	AssCode     string
	StudentID   string
	Data        utils.AssignmentConfig
}

func NewConfigurator() *Configurator {
	config, err := utils.GetConfig()
	var input string

	if err != nil {
		log.Fatalf("Couldn't get config file: %v \n", err)
		return nil
	}

	configData := utils.AssignmentConfig{
		ProjectName: config.ProjectName,
		Directory:   config.Directory,
	}

	if config.AssignmentCode != "" || config.StudentID != "" {
		input, err = utils.ReadInput(fmt.Sprintf("Config already exists with Assignment code: %s and student id: %s. Override? (yes or no) >>", config.AssignmentCode, config.StudentID))
		if err != nil {
			log.Fatal("failed to get override decision")
		}
		if input != "yes" {
			log.Fatalf("Maintaining existing config\n")
			return nil
		}
	}

	return &Configurator{
		Data: configData,
	}
}

func (c *Configurator) ConfigureDirectory() error {
	flags := flag.NewFlagSet("config", flag.ContinueOnError)
	flags.BoolVar(&c.Interactive, "I", false, "Interactive")
	flags.StringVar(&c.AssCode, "code", "", "Quiz code")
	flags.StringVar(&c.StudentID, "student_id", "", "Student ID")
	flags.Parse(os.Args[2:])

	if c.Interactive {
		if err := c.getDataInteractively(); err != nil {
			return err
		}
	}

	if err := c.writeToConfigFile(); err != nil {
		return err
	}

	return nil
}

func (c *Configurator) getDataInteractively() error {
	fmt.Print("Enter your assignment code: ")
	if err := utils.ReadInputUntilValid(&c.AssCode); err != nil {
		return err
	}

	fmt.Print("Enter your student ID: ")
	if err := utils.ReadInputUntilValid(&c.StudentID); err != nil {
		return err
	}

	if c.AssCode == "" || c.StudentID == "" {
		return errors.New("failed to enter assignment code and student ID")
	}

	return nil
}

func (c *Configurator) writeToConfigFile() error {
	if c.AssCode == "" || c.StudentID == "" {
		return errors.New("both assignment code and student ID are required")
	}

	fmt.Printf("Configuring with Assignment Code: %v, Student ID: %v\n", c.AssCode, c.StudentID)

	c.Data.AssignmentCode = c.AssCode
	c.Data.StudentID = c.StudentID

	newFile, _ := json.MarshalIndent(c.Data, "", "")

	err := os.WriteFile(filepath.Join(".subsys", "config.json"), newFile, 0666)

	if err != nil {
		return err
	}

	fmt.Printf("Configuration successful\n")
	return nil
}
