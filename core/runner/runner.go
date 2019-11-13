package runner

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/axetroy/s4/core/grammar"
	"github.com/axetroy/s4/core/ssh"
	"github.com/axetroy/s4/core/variable"
	"github.com/fatih/color"
)

type Runner struct {
	ssh       *ssh.Client       // current ssh client
	step      int               // current step
	cwdLocal  string            // current working dir at local
	tokens    []grammar.Token   // token from parsing
	cwdRemote string            // current remote working dir
	env       map[string]string // env for remote
	variable  map[string]string // var
}

func NewRunner(configFilepath string) (*Runner, error) {
	if f, err := os.Stat(configFilepath); err != nil {
		msg := fmt.Sprintf("Config file `%s` not found", configFilepath)
		return nil, errors.New(color.RedString(msg))
	} else {
		if f.IsDir() {
			msg := fmt.Sprintf("Config file `%s` is not a file", configFilepath)
			return nil, errors.New(color.RedString(msg))
		}
	}

	fmt.Printf("Load the s4 file `%s`\n", color.GreenString(configFilepath))

	content, err := ioutil.ReadFile(configFilepath)

	if err != nil {
		return nil, err
	}

	tokens, err := grammar.Tokenizer(string(content))

	if err != nil {
		return nil, err
	}

	return &Runner{
		tokens:   tokens,
		env:      map[string]string{},
		variable: map[string]string{},
	}, nil
}

func (r *Runner) requireConnection() error {
	if r.ssh == nil {
		return errors.New("you need to connect to server first")
	} else {
		return nil
	}
}

func (r *Runner) resolveLocalPath(localPath string) string {
	if path.IsAbs(localPath) {
		return localPath
	} else {
		return path.Join(r.cwdLocal, localPath)
	}
}

func (r *Runner) resolveLocalPaths(localPaths []string) []string {
	var paths []string

	for _, remotePath := range localPaths {
		paths = append(paths, r.resolveLocalPath(remotePath))
	}

	return paths
}

func (r *Runner) resolveRemotePath(remotePath string) string {
	if path.IsAbs(remotePath) {
		return remotePath
	} else {
		return path.Join(r.cwdRemote, remotePath)
	}
}

func (r *Runner) resolveRemotePaths(remotePaths []string) []string {
	var paths []string

	for _, remotePath := range remotePaths {
		paths = append(paths, r.resolveRemotePath(remotePath))
	}

	return paths
}

func (r *Runner) Run() error {
	defer func() {
		if r.ssh != nil {
			_ = r.ssh.Disconnect()
		}
	}()

	for _, action := range r.tokens {
		r.step++
		switch action.Key {
		case grammar.ActionCONNECT:
			params := action.Node.(grammar.NodeConnect)

			fmt.Printf("[step %v]: CONNECT %s\n", r.step, color.GreenString(fmt.Sprintf("%s@%s:%s", params.Username, params.Host, params.Port)))

			// if ssh client exist. disconnect first
			if r.ssh != nil {
				if err := r.ssh.Disconnect(); err != nil {
					return err
				}
				r.ssh = nil
			}

			password := ""

			if params.Password != nil {
				password = *params.Password
				password = variable.Compile(password, r.variable)
			} else {
				// ask password for remote server
				prompt := &survey.Password{
					Message: "Please type remote server's password",
				}

				if err := survey.AskOne(prompt, &password); err != nil {
					return err
				}
			}

			r.ssh = ssh.NewSSH()

			if err := r.ssh.Connect(params.Host, params.Port, params.Username, password); err != nil {
				r.ssh = nil
				return err
			}

			if cwd, err := os.Getwd(); err != nil {
				return err
			} else {
				r.cwdLocal = cwd
			}

			if remoteCwd, err := r.ssh.Pwd(); err != nil {
				return err
			} else {
				r.cwdRemote = remoteCwd
			}

			break
		case grammar.ActionVAR:
			if err := r.actionVar(action.Node.(grammar.NodeVar)); err != nil {
				return err
			}
			break
		case grammar.ActionENV:
			if err := r.actionEnv(action.Node.(grammar.NodeEnv)); err != nil {
				return err
			}
			break
		case grammar.ActionCD:
			if err := r.actionCd(action.Node.(grammar.NodeCd)); err != nil {
				return err
			}
			break
		case grammar.ActionCMD:
			if err := r.actionCmd(action.Node.(grammar.NodeCmd)); err != nil {
				return err
			}
			break
		case grammar.ActionRUN:
			if err := r.actionRun(action.Node.(grammar.NodeRun)); err != nil {
				return err
			}
			break
		case grammar.ActionMOVE:
			if err := r.actionMove(action.Node.(grammar.NodeCopy)); err != nil {
				return err
			}
			break
		case grammar.ActionCOPY:
			if err := r.actionCopy(action.Node.(grammar.NodeCopy)); err != nil {
				return err
			}
			break
		case grammar.ActionDELETE:
			if err := r.actionDelete(action.Node.(grammar.NodeDelete)); err != nil {
				return err
			}
			break
		case grammar.ActionUPLOAD:
			if err := r.actionUpload(action.Node.(grammar.NodeUpload)); err != nil {
				return err
			}
			break
		case grammar.ActionDOWNLOAD:
			if err := r.actionDownload(action.Node.(grammar.NodeUpload)); err != nil {
				return err
			}
			break
		default:
			return errors.New(fmt.Sprintf("Invalid action `%s`", action.Key))
		}
	}

	r.step++

	fmt.Printf("[step %d]: %s\n", r.step, color.GreenString("done!"))

	return nil
}

func (r *Runner) actionCd(params grammar.NodeCd) error {
	if err := r.requireConnection(); err != nil {
		return err
	}

	dir := params.Target

	fmt.Printf("[step %d]: CD %s\n", r.step, color.GreenString(dir))

	cwd := variable.Compile(dir, r.variable)

	r.cwdRemote = r.resolveRemotePath(cwd)

	return nil
}

func (r *Runner) actionCmd(params grammar.NodeCmd) error {
	fmt.Printf("[step %d]: CMD %s\n", r.step, color.YellowString(fmt.Sprintf("%v", params.SourceCode)))

	command := variable.Compile(params.Command, r.variable)
	args := variable.CompileArray(params.Arguments, r.variable)

	c := exec.Command(command, args...)

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return err
	}

	if c.ProcessState.Success() == false {
		return errors.New(fmt.Sprintf("run command '%s' fail.", params.SourceCode))
	}

	return nil
}

func (r *Runner) actionCopy(params grammar.NodeCopy) error {
	if err := r.requireConnection(); err != nil {
		return err
	}

	sourceFilepath := params.Source
	destinationFilepath := params.Destination

	fmt.Printf("[step %d]: COPY %s to %s\n", r.step, color.YellowString(sourceFilepath), color.GreenString(destinationFilepath))

	sourceFilepath = variable.Compile(sourceFilepath, r.variable)
	destinationFilepath = variable.Compile(destinationFilepath, r.variable)

	sourceFilepath = r.resolveRemotePath(sourceFilepath)
	destinationFilepath = r.resolveRemotePath(destinationFilepath)

	if err := r.ssh.Copy(sourceFilepath, destinationFilepath); err != nil {
		return err
	}
	return nil
}

func (r *Runner) actionDelete(params grammar.NodeDelete) error {
	if err := r.requireConnection(); err != nil {
		return err
	}

	fmt.Printf("[step %v]: DELETE %s\n", r.step, color.YellowString(strings.Join(params.Targets, ",")))

	args := variable.CompileArray(params.Targets, r.variable)

	files := r.resolveRemotePaths(args)

	if err := r.ssh.Delete(files...); err != nil {
		return err
	}

	return nil
}

func (r *Runner) actionDownload(params grammar.NodeUpload) error {
	if err := r.requireConnection(); err != nil {
		return err
	}

	sourceFiles := params.SourceFiles
	destinationDir := params.DestinationDir

	fmt.Printf("[step %d]: DOWNLOAD %s to %s\n", r.step, color.YellowString(strings.Join(sourceFiles, ", ")), color.GreenString(destinationDir))

	sourceFiles = variable.CompileArray(sourceFiles, r.variable)
	destinationDir = variable.Compile(destinationDir, r.variable)

	sourceFiles = r.resolveRemotePaths(sourceFiles)
	destinationDir = r.resolveLocalPath(destinationDir)

	for _, filePath := range sourceFiles {
		if err := r.ssh.Download(filePath, destinationDir); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) actionMove(params grammar.NodeCopy) error {
	if err := r.requireConnection(); err != nil {
		return err
	}

	sourceFilepath := params.Source
	destinationFilepath := params.Destination

	fmt.Printf("[step %d]: MOVE %s to %s\n", r.step, color.YellowString(sourceFilepath), color.GreenString(destinationFilepath))

	sourceFilepath = variable.Compile(sourceFilepath, r.variable)
	destinationFilepath = variable.Compile(destinationFilepath, r.variable)

	sourceFilepath = r.resolveRemotePath(sourceFilepath)
	destinationFilepath = r.resolveRemotePath(destinationFilepath)

	if err := r.ssh.Move(sourceFilepath, destinationFilepath); err != nil {
		return err
	}

	return nil
}

func (r *Runner) actionRun(params grammar.NodeRun) error {
	if err := r.requireConnection(); err != nil {
		return err
	}

	command := params.Command

	fmt.Printf("[step %d]: RUN %s\n", r.step, color.YellowString(command))

	command = variable.Compile(command, r.variable)

	if err := r.ssh.Run(command, ssh.Options{
		CWD: r.cwdRemote,
		Env: r.env,
	}); err != nil {
		return err
	}

	return nil
}

func (r *Runner) actionUpload(params grammar.NodeUpload) error {
	if err := r.requireConnection(); err != nil {
		return err
	}

	sourceFiles := params.SourceFiles
	destinationDir := params.DestinationDir

	fmt.Printf("[step %d]: UPLOAD %s to %s\n", r.step, color.YellowString(strings.Join(sourceFiles, ", ")), color.GreenString(destinationDir))

	sourceFiles = variable.CompileArray(sourceFiles, r.variable)
	destinationDir = variable.Compile(destinationDir, r.variable)

	sourceFiles = r.resolveLocalPaths(sourceFiles)
	destinationDir = r.resolveRemotePath(destinationDir)

	for _, filePath := range sourceFiles {
		if err := r.ssh.Upload(filePath, destinationDir); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) actionEnv(params grammar.NodeEnv) error {
	fmt.Printf("[step %d]: ENV %s\n", r.step, color.GreenString(params.SourceCode))
	r.env[params.Key] = variable.Compile(params.Value, r.variable)
	return nil
}

func (r *Runner) actionVar(params grammar.NodeVar) error {
	fmt.Printf("[step %d]: VAR %s\n", r.step, color.GreenString(params.SourceCode))

	if params.Literal != nil {
		r.variable[params.Key] = params.Literal.Value
	} else if params.Env != nil {
		if params.Env.Local {
			r.variable[params.Key] = os.Getenv(variable.Compile(params.Env.Key, r.variable))
		} else {
			if err := r.requireConnection(); err != nil {
				return err
			}
			if remoteEnvValue, err := r.ssh.Env(variable.Compile(params.Env.Key, r.variable), ssh.Options{Env: r.env}); err != nil {
				return err
			} else {
				r.variable[params.Key] = remoteEnvValue
			}
		}
	} else if params.Command != nil {
		if params.Command.Local {
			commandArr := variable.CompileArray(params.Command.Command, r.variable)

			command := commandArr[0]
			args := commandArr[1:]

			c := exec.Command(command, args...)

			var stdoutBuf bytes.Buffer
			var stderrBuf bytes.Buffer

			c.Stdout = &stdoutBuf
			c.Stderr = &stderrBuf

			if err := c.Run(); err != nil {
				return err
			}

			if c.ProcessState.Success() == false {
				return errors.New(fmt.Sprintf("run command '%s' fail.", params.Command.Command))
			}

			r.variable[params.Key] = strings.TrimSpace(stdoutBuf.String())
		} else {
			if err := r.requireConnection(); err != nil {
				return err
			}

			b, err := r.ssh.RunAndCombineOutput(strings.Join(params.Command.Command, " "), ssh.Options{
				CWD: r.cwdRemote,
				Env: r.env,
			})

			if err != nil {
				return err
			}

			output := string(b)

			r.variable[params.Key] = strings.TrimSpace(output)
		}
	}

	return nil
}
