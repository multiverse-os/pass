package pass

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// PasswordStore represents a password store.
type PasswordStore struct {
	Path    string // path of the store
	GitDir  string // The path of the git directory
	GPGBin  string // The GPG binary to use
	GPGID   string // The GPG ID used to encrypt the passwords
	UsesGit bool   // Whether or not the store uses git
}

// NewPasswordStore returns a new password store.
func NewPasswordStore(storePath string) *PasswordStore {
	s := PasswordStore{}
	s.Path = storePath
	s.UsesGit = true
	s.GitDir = path.Join(s.Path, ".git")

	// Find the GPG bin
	which := exec.Command("which", "gpg2")
	if err := which.Run(); err == nil {
		s.GPGBin = "gpg2"
	} else {
		s.GPGBin = "gpg"
	}

	//Read the .gpg-id file
	gpgIDPath := path.Join(s.Path, ".gpg-id")
	content, _ := ioutil.ReadFile(gpgIDPath)
	s.GPGID = strings.TrimSpace(string(content))

	return &s
}

// Init creates a Password Store at the Path
func (store *PasswordStore) Init(gpgID string) error {
	// Check if the password path already exists
	fi, err := os.Stat(store.Path)
	if err == nil {
		// Path exists, but is it a directory?
		if fi.Mode().IsDir() == false {
			return fmt.Errorf(
				"could not create password store. Path \"%s\" already exists and it is not a directory",
				store.Path)
		}
	} else {
		// Error during os.Stat
		if os.IsNotExist(err) {
			// Path does not exist, create it
			if err := os.Mkdir(store.Path, 0700); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Check if the .gpg-id file already exists.
	gpgIDFilePath := path.Join(store.Path, ".gpg-id")
	fi, err = os.Stat(gpgIDFilePath)
	if err == nil {
		// .gpg-id already exists
		return fmt.Errorf("there is already a .gpg-id file at \"%s\". Stopping init", gpgIDFilePath)
	}

	gpgIDFile, err := os.Create(path.Join(store.Path, ".gpg-id"))
	if err != nil {
		return err
	}
	defer gpgIDFile.Close()
	gpgIDFile.WriteString(gpgID + "\n")
	store.GPGID = gpgID

	if err := store.git("init"); err != nil {
		return err
	}

	return store.AddAndCommit("initial commit", ".gpg-id")
}

// InsertPassword inserts a new password or overwrites an existing one
func (store *PasswordStore) InsertPassword(pwname, pwtext string) error {
	containsPassword, passwordPath := store.ContainsPassword(pwname)

	// Check if password already exists
	var gitAction string
	if containsPassword {
		gitAction = "edited"
	} else {
		gitAction = "added"
	}

	gpg := exec.Command(
		store.GPGBin,
		"-e",
		"-r", store.GPGID,
		"--batch",
		"--use-agent",
		"--no-tty",
		"--yes",
		"-o", passwordPath)

	stdin, _ := gpg.StdinPipe()
	io.WriteString(stdin, pwtext)
	stdin.Close()
	output, err := gpg.CombinedOutput()

	if err != nil {
		return fmt.Errorf("gpg error: \"%s\"", string(output))
	}

	store.AddAndCommit(
		fmt.Sprintf("%s password \"%s\"", gitAction, pwname),
		passwordPath)

	return nil
}

// RemoveDirectory removes the directory at the given path
func (store *PasswordStore) RemoveDirectory(dirname string) error {
	containsDirectory, directoryPath := store.ContainsDirectory(dirname)

	if !containsDirectory {
		return fmt.Errorf("could not find directory at path \"%s\"", directoryPath)
	}

	if err := os.RemoveAll(directoryPath); err != nil {
		return err
	}

	store.AddAndCommit(
		fmt.Sprintf("removed directory \"%s\" from the store", dirname),
		directoryPath)

	return nil
}

// RemovePassword removes the password at the given path
func (store *PasswordStore) RemovePassword(pwname string) error {
	containsPassword, passwordPath := store.ContainsPassword(pwname)

	if !containsPassword {
		return fmt.Errorf("could not find password at path \"%s\"", passwordPath)
	}

	os.Remove(passwordPath)

	store.AddAndCommit(
		fmt.Sprintf("removed password \"%s\" from the store", pwname),
		passwordPath)

	return nil
}

// MoveDirectory moves a directory from source to dest
func (store *PasswordStore) MoveDirectory(source, dest string) error {
	containsDirectory, sourceDirectoryPath := store.ContainsDirectory(source)
	if !containsDirectory {
		return fmt.Errorf("could not find directory at path \"%s\"", sourceDirectoryPath)
	}

	destDirectoryPath := path.Join(store.Path, dest)

	if err := os.Rename(sourceDirectoryPath, destDirectoryPath); err != nil {
		return err
	}

	store.AddAndCommit(
		fmt.Sprintf("moved directory \"%s\" to \"%s\"", source, dest),
		sourceDirectoryPath,
		destDirectoryPath)

	return nil
}

// MovePassword moves a passsword or directory from source to dest.
func (store *PasswordStore) MovePassword(source, dest string) error {
	containsPassword, sourcePasswordPath := store.ContainsPassword(source)

	if !containsPassword {
		return fmt.Errorf("could not find password path \"%s\"", sourcePasswordPath)
	}

	// If the dest ends with a '/', then it is a directory.
	var destPasswordPath string
	if strings.HasSuffix(dest, "/") {
		_, file := filepath.Split(sourcePasswordPath)
		destPasswordPath = path.Join(store.Path, dest, file)
	} else {
		destPasswordPath = path.Join(store.Path, dest+".gpg")
	}

	if err := os.Rename(sourcePasswordPath, destPasswordPath); err != nil {
		return err
	}

	store.AddAndCommit(
		fmt.Sprintf("moved Password \"%s\" to \"%s\"", source, dest),
		sourcePasswordPath,
		destPasswordPath)

	return nil
}

// CopyPassword copies a password from source to dest
func (store *PasswordStore) CopyPassword(source, dest string) error {
	containsPassword, sourcePasswordPath := store.ContainsPassword(source)

	if !containsPassword {
		return fmt.Errorf("could not find password or at path \"%s\"", sourcePasswordPath)
	}

	// If the dest ends with a '/', then it is a directory.
	var destPasswordPath string
	if strings.HasSuffix(dest, "/") {
		_, file := filepath.Split(sourcePasswordPath)
		destPasswordPath = path.Join(store.Path, dest, file)
	} else {
		destPasswordPath = path.Join(store.Path, dest+".gpg")
	}

	if err := CopyFileContents(sourcePasswordPath, destPasswordPath); err != nil {
		return err
	}

	store.AddAndCommit(
		fmt.Sprintf("copied Password \"%s\" to \"%s\"", source, dest),
		destPasswordPath)

	return nil
}

// CopyDirectory copies a directory from source to dest
func (store *PasswordStore) CopyDirectory(source, dest string) error {
	containsDirectory, sourceDirectoryPath := store.ContainsDirectory(source)

	if !containsDirectory {
		return fmt.Errorf("could not find directory at path \"%s\"", path.Join(store.Path, source))
	}

	destDirectoryPath := path.Join(store.Path, dest)
	if err := exec.Command("cp", "-r", sourceDirectoryPath, destDirectoryPath).Run(); err != nil {
		return err
	}

	store.AddAndCommit(
		fmt.Sprintf("copied directory \"%s\" to \"%s\"", source, dest),
		destDirectoryPath)

	return nil
}

// GetPassword returns a decrypted password
func (store *PasswordStore) GetPassword(pwname string) (string, error) {
	containsPassword, passwordPath := store.ContainsPassword(pwname)

	// Error if the password does not exist
	if containsPassword == false {
		return "", fmt.Errorf("could not find password \"%s\" at path \"%s\"", pwname, passwordPath)
	}

	// TODO: Use GPG lib instead
	show := exec.Command(store.GPGBin, "--quiet", "--batch", "--use-agent", "-d", passwordPath)
	output, err := show.CombinedOutput()

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// ContainsPassword returns whether or not the store contains a password with this name.
// it also conveniently returns the password path that was checked
func (store *PasswordStore) ContainsPassword(pwname string) (bool, string) {
	passwordPath := path.Join(store.Path, pwname+".gpg")

	if _, err := os.Stat(passwordPath); os.IsNotExist(err) {
		return false, passwordPath
	}

	return true, passwordPath
}

// ContainsDirectory returns whether or not the store contains a directory with this name.
// it also conveniently returns the directory path that was checked
func (store *PasswordStore) ContainsDirectory(dirname string) (bool, string) {
	directoryPath := path.Join(store.Path, dirname)

	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		return false, directoryPath
	}

	return true, directoryPath
}

// GetPasswordsList returns a list of all the passwords
func (store *PasswordStore) GetPasswordsList() []string {
	var list []string

	var scan = func(path string, fileInfo os.FileInfo, inpErr error) (err error) {
		if strings.HasSuffix(path, ".gpg") {
			_, file := filepath.Split(path)
			password := strings.TrimSuffix(file, ".gpg")
			list = append(list, password)
		}
		return
	}

	filepath.Walk(store.Path, scan)

	return list
}

// AddAndCommit adds paths to the index and creates a commit
func (store *PasswordStore) AddAndCommit(message string, paths ...string) error {
	store.git("reset")

	for _, path := range paths {
		store.git("add", path)
	}

	store.git("commit", "-m", message)

	return nil
}

// git executes a git command
func (store *PasswordStore) git(args ...string) error {
	if !store.UsesGit {
		return nil
	}

	gitArgs := []string{
		"--git-dir=" + store.GitDir,
		"--work-tree=" + store.Path}

	gitArgs = append(gitArgs, args...)

	git := exec.Command("git", gitArgs...)

	// Should we do that?
	git.Stdout = os.Stdout
	git.Stderr = os.Stderr
	git.Stdin = os.Stdin

	if err := git.Run(); err != nil {
		fmt.Println(err.Error())
		return fmt.Errorf("git error: \"%s\"", err.Error())
	}

	return nil
}
