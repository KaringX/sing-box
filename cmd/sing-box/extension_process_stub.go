//go:build with_karing && !(windows || linux)

package main

func makeProcessSingleton() error {
	return nil
}
