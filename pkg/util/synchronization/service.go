package synchronization

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alsritter/middlebaby/pkg/util/logger"
)

type Service struct {
	cfg *Config
	logger.Logger
}

// Config defines the config structure
type Config struct {
	Enable     bool
	StorageDir string
	Repository []*Repository
}

type Repository struct {
	Address string
	Branch  string
}

// 自动同步本地和远程库的服务
func New(cfg *Config, log logger.Logger) (*Service, error) {
	service := &Service{
		cfg:    cfg,
		Logger: log.NewLogger("synchronization"),
	}
	return service, nil
}

// 同步仓库，同步完成后会执行 callback，其中 updated 字段用于通知这个 callback 当前仓库是否有变更过
func (s *Service) Synchronize(ctx context.Context, callback func(repository string, updated bool, err error) error) error {
	for _, repo := range s.cfg.Repository {
		repository := repo.Address
		branch := repo.Branch
		if branch == "" {
			branch = "master"
		}

		location := filepath.Join(s.cfg.StorageDir, getRepositoryPathFromUrl(repository))

		s.Info(map[string]interface{}{
			"location":   location,
			"repository": repository,
			"branch":     branch,
		}, "start to synchronize repository")

		updated, err := s.synchronizeRepository(ctx, repository, branch, location)
		if callback != nil {
			if err := callback(repository, updated, err); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) synchronizeRepository(ctx context.Context, repository string, branch string, location string) (updated bool, err error) {
	_, err = os.Stat(filepath.Join(location, ".git"))
	if err != nil {
		if !os.IsNotExist(err) {
			return false, err
		}
		s.Info(map[string]interface{}{
			"repository": repository,
			"branch":     branch,
			"location":   location,
		}, "start to clone repository")
		if err := cloneRepository(ctx, repository, branch, location); err != nil {
			return false, err
		}
		return true, nil
	}
	s.Info(map[string]interface{}{
		"repository": repository,
		"branch":     branch,
		"location":   location,
	}, "start to get local commits")
	localCommits, err := listRepositoryCommits(ctx, "HEAD", location)
	if err != nil {
		return false, err
	}
	s.Info(map[string]interface{}{
		"repository": repository,
		"location":   location,
		"branch":     branch,
	}, "start to get remote commits")
	remoteCommits, err := listRepositoryCommits(ctx, "origin/"+branch, location)
	if err != nil {
		return false, err
	}
	s.Info(map[string]interface{}{
		"repository":    repository,
		"location":      location,
		"remoteCommits": len(remoteCommits),
		"localCommits":  len(localCommits),
		"branch":        branch,
	}, "start to compare local and remote commits")
	if len(localCommits) == len(remoteCommits) {
		s.Info(map[string]interface{}{
			"repository": repository,
			"location":   location,
			"branch":     branch,
		}, "no need to update local repository")
		return false, nil
	}
	s.Info(map[string]interface{}{
		"repository": repository,
		"location":   location,
		"branch":     branch,
	}, "start to pull repository")
	if err := pullRepository(ctx, repository, branch, location); err != nil {
		return false, err
	}
	s.Info(map[string]interface{}{
		"repository": repository,
		"location":   location,
		"branch":     branch,
	}, "repository updated")
	return true, nil
}

func execCommand(ctx context.Context, args []string, dir string, stdout io.Writer, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	cmd.Dir = dir
	fmt.Println(">", cmd.String())
	return cmd.Run()
}

func cloneRepository(ctx context.Context, address string, branch string, location string) error {
	if branch == "" {
		branch = "master"
	}
	return execCommand(ctx, []string{
		"clone",
		"-b",
		branch,
		address,
		location,
	}, "", os.Stdout, os.Stderr)
}

func pullRepository(ctx context.Context, address string, branch string, location string) error {
	return execCommand(ctx, []string{
		"pull",
		"origin",
		branch,
	}, location, os.Stdout, os.Stderr)
}

func listRepositoryCommits(ctx context.Context, branch string, location string) ([]string, error) {
	var buffer bytes.Buffer
	if err := execCommand(ctx, []string{
		"log",
		branch,
		"--pretty=format:%H",
	}, location, &buffer, os.Stderr); err != nil {
		return nil, err
	}
	commits := strings.Split(buffer.String(), "\n")
	return commits, nil
}

func getRepositoryPathFromUrl(s string) string {
	var projectPath string
	s = strings.TrimSuffix(s, ".git")
	if strings.HasPrefix(s, "git@") {
		parts := strings.Split(s, ":")
		if len(parts) == 2 {
			projectPath = parts[1]
		}
	} else {
		parsed, err := url.Parse(s)
		if err != nil {
			return ""
		}
		projectPath = parsed.Path
	}
	return strings.TrimPrefix(projectPath, "/")
}
