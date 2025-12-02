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
	tagMap := make(map[string]bool)

	for _, tag := range existingTags {
		tagMap[tag] = true
	}

	for _, tag := range newTags {
		tagMap[tag] = true
	}

	result := make([]string, 0, len(tagMap))
	for tag := range tagMap {
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

// RemoveTags removes specified tags from the existing tags list
func RemoveTags(existingTags []string, tagsToRemove []string) []string {
	var result []string

	for _, tag := range existingTags {
		if !ContainsTag(tagsToRemove, tag) {
			result = append(result, tag)
		}
	}

	return result
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
