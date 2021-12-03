package main

import (
	"bufio"
	"fmt"
	"golang.org/x/sys/windows"
	"os"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var DS_STREAM_RENAME = ":wtfbbq"

func openHndl(pwPath *uint16) uintptr {
	//hndl,e := syscall.CreateFile(pwPath,windows.DELETE,0,nil,windows.OPEN_EXISTING,windows.FILE_ATTRIBUTE_NORMAL,0)
	hndl,e := syscall.CreateFile(pwPath,0x00010000,0,nil,3,0x00000080,0)
	if e != nil{
		return 0
	}
	return uintptr(hndl)
}

func renameHndl(hndl uintptr)error{


	type FILE_RENAME_INFO struct {
		Flags          uint32
		RootDirectory  syscall.Handle
		FileNameLength uint32
		FileName       [1]uint16
	}

	//方法1
	/*
		newwinpath := mkwinpathslice(DS_STREAM_RENAME)
		size := int(unsafe.Offsetof(FILE_RENAME_INFO{}.FileName)) + (len(newwinpath))*2
		ibuf := make([]uint8, size)
		info := (*FILE_RENAME_INFO)(unsafe.Pointer(&ibuf[0]))
		info.FileNameLength = uint32(len(newwinpath))*2 - 2
		nbuf := (*[1 << 30]uint16)(unsafe.Pointer(&info.FileName))[:len(newwinpath)]
		copy(nbuf, newwinpath)
		return windows.SetFileInformationByHandle(windows.Handle(hndl),windows.FileRenameInfo, (*byte)(unsafe.Pointer(info)),uint32(size))
	*/

	//方法2
	var fRename FILE_RENAME_INFO
	memset(uintptr(unsafe.Pointer(&fRename)),0,unsafe.Sizeof(fRename))
	lpwStream,_ := syscall.UTF16PtrFromString(DS_STREAM_RENAME)
	fRename.FileNameLength = uint32(unsafe.Sizeof(lpwStream))
	rcmem := syscall.NewLazyDLL(string([]byte{'k','e','r','n','e','l','3','2'})).NewProc(string([]byte{'R','t','l','C','o','p','y','M','e','m','o','r','y'}))
	rcmem.Call(uintptr(unsafe.Pointer(&fRename.FileName)),uintptr(unsafe.Pointer(lpwStream)),unsafe.Sizeof(lpwStream))
	//return windows.SetFileInformationByHandle(windows.Handle(hndl),windows.FileRenameInfo, (*byte)(unsafe.Pointer(&fRename)),uint32(unsafe.Sizeof(fRename)+unsafe.Sizeof(lpwStream)))
	SFIByHandle := syscall.NewLazyDLL(string([]byte{'k','e','r','n','e','l','3','2'})).NewProc(string([]byte{'S','e','t','F','i','l','e','I','n','f','o','r','m','a','t','i','o','n','B','y','H','a','n','d','l','e'}))
	r1,_,e := SFIByHandle.Call(hndl,3, uintptr(unsafe.Pointer(&fRename)),uintptr(unsafe.Sizeof(fRename)+unsafe.Sizeof(lpwStream)))
	if r1 == 0{
		return e
	}
	return nil
}

func mkwinpathslice(path string) []uint16 {
	return utf16.Encode([]rune(path + "\x00"))
}



func depositeHndl(hndl uintptr) error{
	type FILE_DISPOSITION_INFO struct {
		DeleteFile uint32
	}

	fDelete := FILE_DISPOSITION_INFO{}
	memset(uintptr(unsafe.Pointer(&fDelete)),0,unsafe.Sizeof(fDelete))
	fDelete.DeleteFile = 1
	//return windows.SetFileInformationByHandle(windows.Handle(hndl),windows.FileDispositionInfo, (*byte)(unsafe.Pointer(&fDelete)),uint32(unsafe.Sizeof(fDelete)))
	SFIByHandle := syscall.NewLazyDLL(string([]byte{'k','e','r','n','e','l','3','2'})).NewProc(string([]byte{'S','e','t','F','i','l','e','I','n','f','o','r','m','a','t','i','o','n','B','y','H','a','n','d','l','e'}))
	r1,_,e := SFIByHandle.Call(hndl,4, uintptr(unsafe.Pointer(&fDelete)),unsafe.Sizeof(fDelete))
	if r1 == 0{
		return e
	}
	return nil
}

func memset(ptr uintptr, c byte, n uintptr){
	var i uintptr
	for i = 0;i<n;i++{
		pByte:=(*byte)(unsafe.Pointer(ptr+1))
		*pByte = c
	}
}

func main(){
	var wcPath uint16
	memset(uintptr(unsafe.Pointer(&wcPath)),0, unsafe.Sizeof(wcPath))

	windows.GetModuleFileName(0,&wcPath,syscall.MAX_PATH)

	hCurrent := openHndl(&wcPath)
	if hCurrent == ^uintptr(0) || hCurrent == 0{
		fmt.Println("handle err")
		os.Exit(0)
	}
	fmt.Println("open file handler...")


	if renameHndl(hCurrent) != nil {
		fmt.Println("rename err")
		windows.CloseHandle(windows.Handle(hCurrent))
		os.Exit(0)
	}
	fmt.Println("rename file...")

	windows.CloseHandle(windows.Handle(hCurrent))
	fmt.Println("close handler...")


	memset(uintptr(unsafe.Pointer(&wcPath)),0, unsafe.Sizeof(wcPath))
	windows.GetModuleFileName(0,&wcPath,syscall.MAX_PATH)

	hCurrent = openHndl(&wcPath)

	if hCurrent == ^uintptr(0) || hCurrent == 0{
		fmt.Println("handle2 err")
		os.Exit(0)
	}

	fmt.Println("open file handler...")
	e := depositeHndl(hCurrent)
	if e != nil {
		fmt.Println("delete err")
		fmt.Println(e)
		windows.CloseHandle(windows.Handle(hCurrent))
		os.Exit(0)
	}
	fmt.Println("delete file...")

	windows.CloseHandle(windows.Handle(hCurrent))
	fmt.Println("close handler...")

	fmt.Println("Self-Delete Success! ")

	fmt.Print("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

}