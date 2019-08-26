package runner

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey"
	"github.com/axetroy/go-fs"
	"github.com/axetroy/sshunter/lib/parser"
	"github.com/fatih/color"
	"os"
	"path"
	"strings"
)

type Runner struct {
	Config *parser.Config
}

func NewRunner(configFile string) (*Runner, error) {
	if fs.PathExists(configFile) == false {
		msg := fmt.Sprintf("Config file `%s` not found", configFile)
		return nil, errors.New(color.RedString(msg))
	}

	config, err := parser.ParseFile(configFile)

	if err != nil {
		return nil, err
	}

	return &Runner{
		Config: config,
	}, nil
}

func (r *Runner) Run() error {
	if r.Config.Password == "" {
		// ask password for remote server
		password := ""
		prompt := &survey.Password{
			Message: "Please type remote server's password",
		}

		if err := survey.AskOne(prompt, &password); err != nil {
			return err
		}

		r.Config.Password = password
	}

	client := NewSSH(*r.Config)

	localCwd, err := os.Getwd()

	if err != nil {
		return err
	}

	if err := client.Connect(); err != nil {
		return err
	}

	defer client.Disconnect()

	remoteCwd, err := client.Pwd()

	if err != nil {
		return err
	}

	r.Config.CWD = remoteCwd

	step := 1

	for _, action := range r.Config.Actions {

		switch action.Action {
		case "CWD":
			r.Config.CWD = action.Arguments
			fmt.Printf("[Step %v]: CWD %s\n", step, color.GreenString(action.Arguments))
			step += 1
			break
		case "RUN":
			commandWithColor := color.YellowString(fmt.Sprintf("%v", action.Arguments))

			fmt.Printf("[Step %v]: RUN %s\n", step, commandWithColor)

			if err := client.Run(action.Arguments); err != nil {
				return err
			}

			step += 1
			break
		case "MOVE":
			args := strings.Split(action.Arguments, " ")

			if len(args) != 2 {
				return errors.New(fmt.Sprintf("move require source and destination but got `%s`", args))
			}

			sourceFilepath := strings.Trim(args[0], " ")
			destinationFilepath := strings.Trim(args[1], " ")

			if path.IsAbs(sourceFilepath) == false {
				sourceFilepath = path.Join(r.Config.CWD, sourceFilepath)
			}

			if path.IsAbs(destinationFilepath) == false {
				destinationFilepath = path.Join(r.Config.CWD, destinationFilepath)
			}

			fmt.Printf("[Step %v]: MOVE %s to %s\n", step, color.YellowString(sourceFilepath), color.GreenString(destinationFilepath))

			if err := client.Move(sourceFilepath, destinationFilepath); err != nil {
				return err
			}

			step += 1

			break
		case "COPY":
			args := strings.Split(action.Arguments, " ")

			if len(args) != 2 {
				return errors.New(fmt.Sprintf("copy require source and destination but got `%s`", args))
			}

			sourceFilepath := strings.Trim(args[0], " ")
			destinationFilepath := strings.Trim(args[1], " ")

			if path.IsAbs(sourceFilepath) == false {
				sourceFilepath = path.Join(r.Config.CWD, sourceFilepath)
			}

			if path.IsAbs(destinationFilepath) == false {
				destinationFilepath = path.Join(r.Config.CWD, destinationFilepath)
			}

			fmt.Printf("[Step %v]: COPY %s to %s\n", step, color.YellowString(sourceFilepath), color.GreenString(destinationFilepath))

			if err := client.Copy(sourceFilepath, destinationFilepath); err != nil {
				return err
			}

			step += 1

			break
		case "DELETE":
			args := strings.Split(action.Arguments, " ")

			var files []string

			for _, file := range args {
				if path.IsAbs(file) == false {
					file = path.Join(r.Config.CWD, file)
				}

				files = append(files, file)
			}

			fmt.Printf("[Step %v]: DELETE %s\n", step, color.YellowString(action.Arguments))

			if err := client.Delete(files...); err != nil {
				return err
			}

			step += 1

			break
		case "UPLOAD":
			f, err := parser.FileParser(action.Arguments)

			if err != nil {
				return err
			}

			if path.IsAbs(f.Destination) == false {
				if r.Config.CWD != "" {
					f.Destination = path.Join(r.Config.CWD, f.Destination)
				}
			}

			fmt.Printf("[Step %v]: UPLOAD local:%s to remote:%s\n", step, color.YellowString(strings.Join(f.Source, ", ")), color.GreenString(f.Destination))

			for _, filePath := range f.Source {

				if path.IsAbs(filePath) == false {
					filePath = path.Join(localCwd, filePath)
				}

				err := client.Copy(filePath, f.Destination)

				fmt.Println("copy", filePath, "-->", f.Destination)

				if err != nil {
					return err
				}
			}

			step += 1

			break
		case "DOWNLOAD":
			f, err := parser.FileParser(action.Arguments)

			if err != nil {
				return err
			}

			if path.IsAbs(f.Destination) == false {
				f.Destination = path.Join(localCwd, f.Destination)
			}

			fmt.Printf("[Step %v]: DOWNLOAD remote:%s to local:%s\n", step, color.YellowString(strings.Join(f.Source, ", ")), color.GreenString(f.Destination))

			for _, filePath := range f.Source {

				if path.IsAbs(filePath) == false {
					if r.Config.CWD != "" {
						filePath = path.Join(r.Config.CWD, filePath)
					}
				}

				err := client.Download(filePath, f.Destination)

				fmt.Println("download", filePath, "-->", f.Destination)

				if err != nil {
					return err
				}
			}

			step += 1

			break
		default:
			return errors.New(fmt.Sprintf("Invalid action `%s`", action.Action))
		}
	}

	return nil
}