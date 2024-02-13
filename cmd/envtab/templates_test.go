package envtab

import (
	"fmt"
	"testing"
)

func TestMakeLoadoutFromTemplate(t *testing.T) {
	loadout := MakeLoadoutFromTemplate("aws")
	fmt.Printf("DEBUG: Loadout: %v\n", loadout)
}
