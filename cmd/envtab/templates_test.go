package envtab

import (
	"fmt"
	"os"
	"testing"
)

func TestMakeLoadoutFromTemplate(t *testing.T) {

	if _, err := os.Stat(InitEnvtab("") + "/templates"); os.IsNotExist(err) {
		err := os.Mkdir(InitEnvtab("")+"/templates", 0755)
		if err != nil {
			t.Errorf("Error creating test template: %v", err)
		}
	}

	f, err := os.Create(InitEnvtab("") + "/templates/test.yml")
	if err != nil {
		t.Errorf("Error creating test template: %v", err)
	}

	_, err = f.WriteString("entries:\n- AWS_ACCESS_KEY_ID\n")
	if err != nil {
		t.Errorf("Error writing test template: %v", err)
	}

	loadout := MakeLoadoutFromTemplate("test", false)
	fmt.Printf("DEBUG: Loadout: %v\n", loadout)

	// cleanup
	err = os.Remove(InitEnvtab("") + "/templates/test.yml")
}
