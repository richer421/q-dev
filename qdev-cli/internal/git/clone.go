package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
)

const DefaultRepo = "https://github.com/richer421/q-dev"

type CloneOptions struct {
	RepoURL  string
	Tag      string
	GitToken string
	SSHKey   string
}

func Clone(targetDir string, opts CloneOptions) error {
	repoURL := opts.RepoURL
	if repoURL == "" {
		repoURL = DefaultRepo
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	cloneOpts := &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	}

	auth, err := getAuth(repoURL, opts.GitToken, opts.SSHKey)
	if err != nil {
		return fmt.Errorf("认证失败: %w", err)
	}
	if auth != nil {
		cloneOpts.Auth = auth
	}

	if opts.Tag != "" {
		return cloneWithTag(targetDir, repoURL, opts.Tag, auth)
	}

	_, err = git.PlainClone(targetDir, false, cloneOpts)
	if err != nil {
		return fmt.Errorf("克隆仓库失败: %w", err)
	}

	return nil
}

func cloneWithTag(targetDir, repoURL, tag string, auth transport.AuthMethod) error {
	remote := git.NewRemote(memory.NewStorage(), &gitconfig.RemoteConfig{
		Name: "origin",
		URLs: []string{repoURL},
	})

	listOpts := &git.ListOptions{}
	if auth != nil {
		listOpts.Auth = auth
	}

	refs, err := remote.List(listOpts)
	if err != nil {
		return fmt.Errorf("获取远程引用失败: %w", err)
	}

	var tagHash plumbing.Hash
	for _, ref := range refs {
		if ref.Name().Short() == tag {
			tagHash = ref.Hash()
			break
		}
	}

	if tagHash.IsZero() {
		return fmt.Errorf("找不到 tag: %s", tag)
	}

	cloneOpts := &git.CloneOptions{
		URL:           repoURL,
		Progress:      os.Stdout,
		ReferenceName: plumbing.NewTagReferenceName(tag),
		SingleBranch:  true,
		Depth:         1,
	}
	if auth != nil {
		cloneOpts.Auth = auth
	}

	_, err = git.PlainClone(targetDir, false, cloneOpts)
	if err != nil {
		return fmt.Errorf("克隆仓库失败: %w", err)
	}

	return nil
}

func getAuth(repoURL, gitToken, sshKey string) (transport.AuthMethod, error) {
	if sshKey != "" {
		return getSSHAuth(sshKey)
	}

	if gitToken != "" {
		return &http.BasicAuth{
			Username: "git",
			Password: gitToken,
		}, nil
	}

	if len(repoURL) > 0 && repoURL[0] != 'h' {
		defaultKey := filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
		if _, err := os.Stat(defaultKey); err == nil {
			return getSSHAuth(defaultKey)
		}
	}

	return nil, nil
}

func getSSHAuth(keyPath string) (transport.AuthMethod, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("读取 SSH key 失败: %w", err)
	}

	signer, err := gitssh.NewPublicKeys("git", key, "")
	if err != nil {
		return nil, fmt.Errorf("解析 SSH key 失败: %w", err)
	}

	return signer, nil
}
