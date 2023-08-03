// Package alslgr defines Logger with internal buffer in order to minimize amount of expensive function calls,
// e.g. opening/closing file. Also, when these calls are performed, internal buffer if substituting with additional
// one, so writing to Logger isn't blocked while old buffer being dumped.
//
// Package also defines Dumper interface, that allows you to dump buffered data wherever you want.
package alslgr
