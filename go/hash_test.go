package redundancy

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"
	"testing"
	"time"
)

func TestHash(t *testing.T) {
	CreateFixedFile(4*1024*1024*1024, path.Join(".", "test"))
	start := time.Now()

	hashResult, size, err := ComputerHashFromFile("test", 16*1024*1024, 6)

	if err != nil {
		t.Errorf(err.Error())
	}
	fmt.Println("hashResult1:", hashResult)
	fmt.Println("size1 :", size)

	fmt.Println("cost time1:", time.Since(start).Milliseconds())

	start = time.Now()

	hashResul2, size2, err := ComputerHashFromFile("test", 16*1024*1024, 6)

	if err != nil {
		t.Errorf(err.Error())
	}
	fmt.Println("hashResult1:", hashResul2)
	fmt.Println("size1 :", size2)

	fmt.Println("cost time2:", time.Since(start).Milliseconds())
}

// create test file
func CreateFixedFile(size uint64, fileName string) {
	exist, err := PathExists(fileName)
	if exist {
		return
	}
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, size)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	_, err = file.WriteString(string(b))
	if err != nil {
		log.Fatal(err)
	}
}
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
