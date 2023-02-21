// Package alslgr defines Logger interface which is an abstraction of logger having internal buffer
// in order to minimize amount of expensive function calls, e.g. opening/closing file.
//
// Package also defines Dumper interface, that allows you to dump buffered data wherever you want.
package alslgr
