package pui

import (
	"errors"
	"fmt"
	"sort"
	"io/ioutil"
	"encoding/base64"

	"github.com/TNK-Studio/gortal/config"
	"github.com/TNK-Studio/gortal/utils"
	"github.com/TNK-Studio/gortal/utils/logger"
	"github.com/elfgzp/promptui"
	"github.com/elfgzp/ssh"
)

// CreateUser new user
func CreateUser(showAdminSelect bool, isAdmin bool, sess *ssh.Session) (*string, *config.User, error) {
	logger.Logger.Info("Create a user.")
	stdio := utils.SessIO(sess)
	usernamePui := promptui.Prompt{
		Label: "New Username",
		Validate: MultiValidate([](func(string) error){
			func(input string) error {
				if len(input) < 3 {
					return errors.New("Username must have more than 3 characters")
				}
				return nil
			},
			func(input string) error {
				user := config.Conf.GetUserByUsername(input)
				if user != nil {
					return fmt.Errorf("Username '%s' of user is existed", input)
				}
				return nil
			},
		}),
		Stdin:  stdio,
		Stdout: stdio,
	}

	username, err := usernamePui.Run()
	if err != nil {
		return nil, nil, err
	}

	passwdPui := promptui.Prompt{
		Label: "Password",
		Validate: func(input string) error {
			if len(input) < 6 {
				return errors.New("Password must have more than 6 characters")
			}
			return nil
		},
		Mask:   '*',
		Stdin:  stdio,
		Stdout: stdio,
	}

	passwd, err := passwdPui.Run()
	if err != nil {
		return nil, nil, err
	}

	confirmPasswdPui := promptui.Prompt{
		Label: "Confirm your password",
		Validate: func(input string) error {
			if input != passwd {
				return errors.New("Password not match")
			}
			return nil
		},
		Mask:   '*',
		Stdin:  stdio,
		Stdout: stdio,
	}

	_, err = confirmPasswdPui.Run()
	if err != nil {
		return nil, nil, err
	}

	publicKeyPui := promptui.Prompt{
		Label: "Your publicKey's position: ",
		Validate: MultiValidate([]func(string) error{
			func(input string) error {
				if !utils.FileExited(input) {
					return errors.New("File not found")
				}
				return nil
			},
		}),
		Stdin:  stdio,
		Stdout: stdio,
	}

	publicKeyPos, err := publicKeyPui.Run()
	if err != nil {
		return nil, nil, err
	}
	publicKeyFile, err := ioutil.ReadFile(utils.FilePath(publicKeyPos))
	if err != nil {
		logger.Logger.Warningf("Error reading publicKey file: %s\n", err)
		return nil, nil, err
	}
	publicKeyBase64:= base64.StdEncoding.EncodeToString(publicKeyFile)

	IsAdminString := ""
	if showAdminSelect && !isAdmin {
		adminPui := promptui.Prompt{
			Label:    "Is admin ? yes/no",
			Validate: YesOrNo(),
			Stdin:    stdio,
			Stdout:   stdio,
		}

		IsAdminString, err = adminPui.Run()
		if err != nil {
			return nil, nil, err
		}
	}

	isAdmin = IsAdminString == "yes" || isAdmin
	if isAdmin {
		logger.Logger.Info("Create a admin user")
	}
	key, user := config.Conf.AddUser(username, passwd, isAdmin, publicKeyBase64)
	return &key, user, nil
}

// GetUsersMenu get users
func GetUsersMenu(selectedFunc func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) error) func(int, *MenuItem, *ssh.Session, []*MenuItem) *[]*MenuItem {
	return func(index int, menuItem *MenuItem, sess *ssh.Session, selectedChain []*MenuItem) *[]*MenuItem {
		menu := make([]*MenuItem, 0)
		userKeys := make([]string, 0)
		for userKey := range *config.Conf.Users {
			userKeys = append(userKeys, userKey)
		}
		sort.Strings(userKeys)
		if len(userKeys) < 1 {
			return &menu
		}
		for _, userKey := range userKeys {
			user := (*config.Conf.Users)[userKey]
			info := make(map[string]string, 0)
			info[userInfoKey] = userKey
			menu = append(
				menu,
				&MenuItem{
					Label:             user.Username,
					Info:              info,
					SubMenuTitle:      fmt.Sprintf("Please select. "),
					SelectedFunc:      selectedFunc,
					BackAfterSelected: true,
				},
			)
		}
		return &menu
	}
}

// ChangePassword ChangePassword\
func ChangePassword(username string, sess *ssh.Session) error {
	logger.Logger.Infof("GetChangePassword of user '%s'.", username)
	stdio := utils.SessIO(sess)

	user := (*config.Conf).GetUserByUsername(username)
	if user == nil {
		return fmt.Errorf("Username '%s' of user not existed. ", username)
	}

	passwdPui := promptui.Prompt{
		Label: "Password",
		Validate: func(input string) error {
			if len(input) < 6 {
				return errors.New("Password must have more than 6 characters")
			}
			return nil
		},
		Mask:   '*',
		Stdin:  stdio,
		Stdout: stdio,
	}

	passwd, err := passwdPui.Run()
	if err != nil {
		return err
	}

	confirmPasswdPui := promptui.Prompt{
		Label: "Confirm your password",
		Validate: func(input string) error {
			if input != passwd {
				return errors.New("Password not match")
			}
			return nil
		},
		Mask:   '*',
		Stdin:  stdio,
		Stdout: stdio,
	}

	_, err = confirmPasswdPui.Run()
	if err != nil {
		return err
	}
	// Todo Hash password
	user.HashPasswd = passwd
	return nil
}
