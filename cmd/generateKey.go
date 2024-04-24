/*
Copyright © 2024 Stephen Owens steve.owens@rightfoot.consulting
*/
package cmd

import (
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rightfoot-consulting/p2pbbs/bbscrypto"
	"github.com/spf13/cobra"
)

// generateKeyCmd represents the generateKey command
var generateKeyCmd = &cobra.Command{
	Use:   "generateKey",
	Short: "Generate a keypair for use with libp2p",
	Long:  `This command can generate public private keypairs and stores the private key in a file while printing out a public key.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("generateKey called")
		keygenerator, err := newKeyGenerator(cmd)
		if err != nil {
			panic(err)
		}
		keygenerator.generateKey()
	},
}

func init() {
	rootCmd.AddCommand(generateKeyCmd)
	generateKeyCmd.Flags().StringP("type", "t", "ed25519", "Specify the type of keypair to generate defaults to 'ed25519', can also be: 'ecdsa', 'rsa' or 'secp256k1'")
	generateKeyCmd.Flags().StringP("out", "o", "private.key", "Specify the file path for the generated key, default is 'private.key' in current working directory")
	generateKeyCmd.Flags().Int32P("bits", "b", 0, "For RSA keys specifies the bit size of the key can be: 1024, 2048 or 4096, defaults to 2048")
	generateKeyCmd.Flags().StringP("curve", "c", "p256", "For ecdsa curves specifies the curve to use can be:  'p224', 'p256', 'p384', or 'p521', defaults to 'p256'")
}

type keytype string

const (
	ed25519        keytype = "ed25519"
	ecdsa          keytype = "ecdsa"
	rsa            keytype = "rsa"
	secp256k1      keytype = "secp256k1"
	invalidKeytype keytype = "invalid"
)

type ecdsaCurve string

const (
	p224              ecdsaCurve = "p224"
	p256              ecdsaCurve = "p256"
	p384              ecdsaCurve = "p384"
	p521              ecdsaCurve = "p521"
	invalidEcdsaCurve            = "invalid"
)

func parseKeytype(val string) (keytype, error) {
	switch val {
	case "ed25519":
		return ed25519, nil
	case "ecdsa":
		return ecdsa, nil
	case "rsa":
		return rsa, nil
	case "secp256k1":
		return secp256k1, nil
	default:
		return invalidKeytype, fmt.Errorf("invalid keytype %s", val)
	}
}
func parseEcdsaCurve(val string) (ecdsaCurve, error) {
	switch val {
	case "p224":
		return p224, nil
	case "p256":
		return p256, nil
	case "p384":
		return p384, nil
	case "p521":
		return p521, nil
	default:
		return invalidEcdsaCurve, fmt.Errorf("invalid ecdsa curve %s", val)
	}
}

type keygenerator struct {
	keyType keytype
	outPath string
	bits    int32
	curve   ecdsaCurve
}

func newKeyGenerator(cmd *cobra.Command) (generator *keygenerator, err error) {
	var curve ecdsaCurve
	ktParam, err := cmd.Flags().GetString("type")
	if err != nil {
		return
	}
	bits, err := cmd.Flags().GetInt32("bits")
	if err != nil {
		return
	}
	curveParam, err := cmd.Flags().GetString("curve")
	if err != nil {
		return
	}
	outpath, err := cmd.Flags().GetString("out")
	if err != nil {
		return
	}
	if outpath == "" {
		outpath = "private.key"
	}

	kt, err := parseKeytype(ktParam)
	if err != nil {
		return
	}
	if kt == ecdsa {
		curve, err = parseEcdsaCurve(curveParam)
		if err != nil {
			return
		}
	}
	if kt == rsa {
		if bits == 0 {
			bits = 2048
		}
		if !(bits == 1024 || bits == 2048 || bits == 4096) {
			err = fmt.Errorf("invalid bit size %d for rsa key", bits)
			return
		}
	}
	generator = &keygenerator{
		keyType: kt,
		outPath: outpath,
		bits:    bits,
		curve:   curve,
	}
	return
}

func (kg *keygenerator) generateKey() {
	switch kg.keyType {
	case ecdsa:
		kg.generateEcdsaKey()
	case ed25519:
		kg.generateNonEcdsaKey(crypto.Ed25519)
	case rsa:
		kg.generateNonEcdsaKey(crypto.RSA)
	case secp256k1:
		kg.generateNonEcdsaKey(crypto.Secp256k1)
	default:
		panic(fmt.Errorf("unsupported key type %v", kg.keyType))
	}
}

func (kg *keygenerator) generateEcdsaKey() {
	var curve elliptic.Curve
	var reader io.Reader = rand.Reader
	switch kg.curve {
	case p224:
		curve = elliptic.P224()
	case p256:
		curve = elliptic.P256()
	case p384:
		curve = elliptic.P384()
	case p521:
		curve = elliptic.P521()
	default:
		panic(fmt.Errorf("unsupported elliptic curve type %v", kg.curve))
	}
	privateKey, _, err := crypto.GenerateECDSAKeyPairWithCurve(curve, reader)
	if err != nil {
		panic(err)
	}
	kg.outputKeys(privateKey)
}

func (kg *keygenerator) generateNonEcdsaKey(kt int) {
	var reader io.Reader = rand.Reader
	privateKey, _, err := crypto.GenerateKeyPairWithReader(kt, int(kg.bits), reader)
	if err != nil {
		panic(err)
	}
	kg.outputKeys(privateKey)
}

func (kg *keygenerator) outputKeys(privateKey crypto.PrivKey) {
	id, err := peer.IDFromPrivateKey(privateKey)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nSaving private key for node with ID: %s, to: %s", id.String(), kg.outPath)
	err = bbscrypto.SavePrivateKey(kg.outPath, privateKey)
	if err != nil {
		panic(err)
	}
}
