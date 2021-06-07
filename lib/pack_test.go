/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package swamppack_test

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"testing"

	swamppack "github.com/swamp/pack/lib"
)

func Test(t *testing.T) {
	repo := swamppack.NewConstantRepo()

	funcDecl := repo.AddFunctionDeclaration("someFunction", 2, 1)

	repo.AddExternalFunction("coreListMap", 4)

	repo.AddBoolean(false)

	repo.AddInteger(48)
	repo.AddInteger(50)

	repo.AddString("can you see this?")

	constants := []*swamppack.Constant{funcDecl}

	opcodes := []byte{48, 49, 50, 51, 52, 53}

	repo.AddFunction("functionWithOpcodes", 3, 3, 2, constants, opcodes)

	octets, err := swamppack.Pack(repo, nil)
	if err != nil {
		fmt.Printf("%v", err)
	}

	encoded := hex.EncodeToString(octets)

	fmt.Printf("hex:%v\n", encoded)
	ioutil.WriteFile("test.swamp-pack", octets, 0o644)

	const expected = `f09fa68a524146460af09f93a673706b3400000000f09f939c7374693000000000f09f92bb7363643000000071f09f91be01040b636f72654c6973744d61700000f09f9b8200000001010c736f6d6546756e6374696f6e0002f09f909c0100f09f94a2020000003000000032f09f8ebb011163616e20796f752073656520746869733ff09f8cb300f09f908a000000010302000100010006303132333435`
	if encoded != expected {
		t.Errorf("wrong encoding")
	}
}
