package envtab

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	tagz "github.com/gmherb/envtab/pkg/tags"
	"github.com/gmherb/envtab/pkg/utils"
	yaml "gopkg.in/yaml.v2"
)

type LoadoutMetadata struct {
	CreatedAt   string   `json:"createdAt" yaml:"createdAt"`
	LoadedAt    string   `json:"loadedAt" yaml:"loadedAt"`
	UpdatedAt   string   `json:"updatedAt" yaml:"updatedAt"`
	Login       bool     `json:"login" yaml:"login"`
	Tags        []string `json:"tags" yaml:"tags"`
	Description string   `json:"description" yaml:"description"`
}

type Loadout struct {
	Metadata LoadoutMetadata   `json:"metadata" yaml:"metadata"`
	Entries  map[string]string `json:"entries" yaml:"entries"`
}

func (l Loadout) Export() {

	pathMap := make(map[string]bool)
	order := []string{}

	for _, p := range strings.Split(os.Getenv("PATH"), string(os.PathListSeparator)) {
		if _, exists := pathMap[p]; !exists {
			order = append(order, p)
		}
		pathMap[p] = true
	}

	re := regexp.MustCompile(`\$PATH`)
	reEnc := regexp.MustCompile(`^ENC:`)

	for key, value := range l.Entries {
		if value != "" {
			match := re.MatchString(value)
			encrypted := reEnc.MatchString(value)
			if key == "PATH" && match {

				newPath := re.ReplaceAllString(value, "")
				newPath = strings.Trim(newPath, ":")

				println("DEBUG: Found potential new PATH(s) [" + newPath + "].")

				for _, np := range strings.Split(newPath, string(os.PathListSeparator)) {
					if _, exists := pathMap[np]; !exists {

						println("DEBUG: Adding new path [" + np + "] to PATH map.")
						order = append(order, np)
						pathMap[np] = true
					}
				}

				paths := make([]string, 0, len(pathMap)) // preallocate for efficiency
				// preserve order with loop
				for _, k := range order {
					paths = append(paths, k)
				}

				os.Setenv("PATH", strings.Join(paths, string(os.PathListSeparator)))
				fmt.Printf("export PATH=%s\n", os.Getenv("PATH"))
			} else if encrypted {
				// TODO DECRYPYT
				println("DEBUG: value for [" + key + "] is encrypted.")
				fmt.Printf("export %s=%s\n", key, value)
			} else {
				fmt.Printf("export %s=%s\n", key, value)
			}
		}
	}
	l.UpdateLoadedAt()
}

func (l *Loadout) UpdateEntry(key string, value string) error {
	println("DEBUG: UpdateEntry called")
	l.Entries[key] = value
	l.UpdateUpdatedAt()
	return nil
}

func (l *Loadout) UpdateTags(tags []string) error {
	println("DEBUG: UpdateTags called")
	l.Metadata.Tags = tagz.MergeTags(l.Metadata.Tags, tags)
	l.UpdateUpdatedAt()
	return nil
}

func (l *Loadout) ReplaceTags(tags []string) error {
	println("DEBUG: ReplaceTags called")
	l.Metadata.Tags = tags
	l.UpdateUpdatedAt()
	return nil
}

func (l *Loadout) UpdateDescription(description string) error {
	println("DEBUG: UpdateDescription called")
	l.Metadata.Description = description
	l.UpdateUpdatedAt()
	return nil
}

func (l *Loadout) UpdateLogin(login bool) error {
	println("DEBUG: UpdateLogin called")
	l.Metadata.Login = login
	l.UpdateUpdatedAt()
	return nil
}

func (l *Loadout) UpdateUpdatedAt() error {
	println("DEBUG: UpdateUpdatedAt called")
	l.Metadata.UpdatedAt = utils.GetCurrentTime()
	return nil
}

func (l *Loadout) UpdateLoadedAt() error {
	println("DEBUG: UpdateLoadedAt called")
	l.Metadata.LoadedAt = utils.GetCurrentTime()
	return nil
}

func (l *Loadout) PrintLoadout() error {

	data, err := yaml.Marshal(l)
	if err != nil {
		return err
	}

	fmt.Printf("%s", string(data))

	return nil
}

// Initialize a new Loadout struct
func InitLoadout() *Loadout {

	loadout := &Loadout{
		Metadata: LoadoutMetadata{
			CreatedAt:   utils.GetCurrentTime(),
			LoadedAt:    utils.GetCurrentTime(),
			UpdatedAt:   utils.GetCurrentTime(),
			Login:       false,
			Tags:        []string{},
			Description: "",
		},
		Entries: map[string]string{},
	}

	return loadout
}

func CompareLoadouts(old Loadout, new Loadout) bool {
	if old.Metadata.CreatedAt != new.Metadata.CreatedAt {
		return true
	}
	if old.Metadata.LoadedAt != new.Metadata.LoadedAt {
		return true
	}
	if old.Metadata.UpdatedAt != new.Metadata.UpdatedAt {
		return true
	}
	if old.Metadata.Login != new.Metadata.Login {
		return true
	}
	if len(old.Metadata.Tags) != len(new.Metadata.Tags) {
		return true
	}
	for i, tag := range old.Metadata.Tags {
		if tag != new.Metadata.Tags[i] {
			return true
		}
	}
	if old.Metadata.Description != new.Metadata.Description {
		return true
	}
	if len(old.Entries) != len(new.Entries) {
		return true
	}
	for key, value := range old.Entries {
		if value != new.Entries[key] {
			return true
		}
	}
	return false
}
