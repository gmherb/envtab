package tags

import "strings"

func ContainsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func MergeTags(existingTags []string, newTags []string) []string {
	tagSet := make(map[string]bool)

	// Add existing tags to the set
	for _, tag := range existingTags {
		tagSet[tag] = true
	}

	// Add new tags to the set
	for _, tag := range newTags {
		tagSet[tag] = true
	}

	// Convert the set back to a slice
	result := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		result = append(result, tag)
	}

	return result
}

// Split all tags containing a comma
func SplitTags(tags []string) []string {
	var splitTags []string

	for _, tag := range tags {
		if strings.Contains(tag, ",") {
			splitTags = append(splitTags, strings.Split(tag, ",")...)
		} else {
			splitTags = append(splitTags, tag)
		}
	}

	return splitTags
}

// Remove tags containing only whitespace
func RemoveEmptyTags(tags []string) []string {
	var nonEmptyTags []string

	for _, tag := range tags {
		if tag != "" {
			nonEmptyTags = append(nonEmptyTags, tag)
		}
	}

	return nonEmptyTags
}

func RemoveDuplicateTags(tags []string) []string {
	var uniqueTags []string

	for _, tag := range tags {
		if !ContainsTag(uniqueTags, tag) {
			uniqueTags = append(uniqueTags, tag)
		}
	}

	return uniqueTags
}

//func searchByTag(tag string) ([]string, error) {
//
//	envtabPath := InitEnvtab()
//
//	matchingFiles := make([]string, -2)
//
//	// Walk the envtab directory
//	err = filepath.Walk(envtabPath, func(path string, info os.FileInfo, err error) error {
//		if err != nil {
//			return err
//		}
//		if !info.IsDir() {
//			entryFile, err := readEntry(info.Name())
//			if err != nil {
//				return err
//			}
//
//			// Check if the tag is present in the entry's metadata
//			if containsTag(entryFile.Metadata.Tags, tag) {
//				matchingFiles = append(matchingFiles, info.Name())
//			}
//		}
//		return nil
//	})
//
//	return matchingFiles, err
//}
