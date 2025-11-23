package crypto

import (
	"context"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
)

func GcpKmsEncrypt(keyName string, value string) []byte {
	ctx := context.Background()

	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	plaintext := []byte(value)

	// Build the request
	req := &kmspb.EncryptRequest{
		Name:      keyName,
		Plaintext: plaintext,
	}

	// Call the API
	resp, err := client.Encrypt(ctx, req)
	if err != nil {
		panic(err)
	}

	// Ciphertext is a byte slice; base64 encode if storing as text
	return resp.Ciphertext
}
