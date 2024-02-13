package envtab

import (
	"fmt"
	"os"
	"testing"
)

func TestMakeLoadoutFromTemplate(t *testing.T) {

	name := "TestMakeLoadoutFromTemplate"
	templateDir := InitEnvtab("") + "/templates"
	filepath := templateDir + "/" + name + ".yml"

	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		err := os.Mkdir(templateDir, 0755)
		if err != nil {
			t.Errorf("Error creating %s template: %v", filepath, err)
		}
	}

	f, err := os.Create(filepath)
	if err != nil {
		t.Errorf("Error creating %s template: %v", filepath, err)
	}

	_, err = f.WriteString("entries:\n- AWS_ACCESS_KEY_ID\n")
	if err != nil {
		t.Errorf("Error writing %s template: %v", filepath, err)
	}

	loadout := MakeLoadoutFromTemplate(name, false)
	fmt.Printf("DEBUG: Loadout: %v\n", loadout)

	if loadout.Entries["AWS_ACCESS_KEY_ID"] != "" {
		t.Errorf("Error creating %s template: %v", filepath, "AWS_ACCESS_KEY_ID should be empty")
	}

	// cleanup
	err = os.Remove(filepath)
}
