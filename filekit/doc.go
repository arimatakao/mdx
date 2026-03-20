// Package filekit provides abstractions for writing manga chapter pages into
// different output containers such as CBZ, PDF, EPUB, or a plain directory.
//
// A typical flow is:
//  1. Create a container with NewContainer.
//  2. Add pages with Container.AddFile.
//  3. Finalize output with Container.WriteOnDiskAndClose.
package filekit
