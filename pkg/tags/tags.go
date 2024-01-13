package tags

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
