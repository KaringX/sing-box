//go:build !(windows || linux)

// karing
package main

func makeProcessSingleton() error {
	return nil
}
