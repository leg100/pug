package app

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	// must parse flags before calling testing.Short()
	flag.Parse()
	if testing.Short() {
		return
	}
	// Remove all .terraform directories from testdata. These can sometimes be
	// created outside of tests, e.g. running pug within the repo, but none the
	// of tests should start with a .terraform directory, otherwise it can lead
	// to tests falsely passing.
	{
		dirs, _ := filepath.Glob("./testdata/*/modules/*/.terraform")
		for _, dir := range dirs {
			os.RemoveAll(dir)
		}
	}
	// Remove all .terraform.lock.hcl files from testdata. These can sometimes be
	// created outside of tests, e.g. running pug within the repo, but none the
	// of tests should start with a .terraform.lock.hcl directory, otherwise it
	// can cause unintended consequences.
	{
		files, _ := filepath.Glob("./testdata/*/modules/*/.terraform.lock.hcl")
		for _, file := range files {
			os.Remove(file)
		}
	}
	// Remove all .terraform-cache directories from testdata. These can sometimes be
	// created outside of tests, e.g. running pug within the repo, but none the
	// of tests should start with a .terraform-cache directory, otherwise it
	// can cause unintended consequences.
	{
		dirs, _ := filepath.Glob("./testdata/*/modules/*/.terraform-cache")
		for _, dir := range dirs {
			os.RemoveAll(dir)
		}
	}

	os.Exit(m.Run())
}
