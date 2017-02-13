package main

import (
	"github.com/GeertJohan/go.rice/embedded"
	"time"
)

func init() {

	// define files
	file2 := &embedded.EmbeddedFile{
		Filename:    `ascii_art.txt`,
		FileModTime: time.Unix(1486922764, 0),
		Content:     string([]byte{0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x20, 0x5f, 0x5f, 0x5f, 0x5f, 0x20, 0x20, 0x5f, 0x5f, 0x20, 0x5f, 0x5f, 0x20, 0x20, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0xa, 0x5c, 0x5f, 0x5f, 0x5f, 0x20, 0x20, 0x20, 0x2f, 0x2f, 0x20, 0x5f, 0x5f, 0x20, 0x5c, 0x7c, 0x20, 0x20, 0x7c, 0x20, 0x20, 0x5c, 0x2f, 0x20, 0x20, 0x5f, 0x5f, 0x5f, 0x2f, 0xa, 0x20, 0x2f, 0x20, 0x20, 0x20, 0x20, 0x2f, 0x5c, 0x20, 0x20, 0x5f, 0x5f, 0x5f, 0x2f, 0x7c, 0x20, 0x20, 0x7c, 0x20, 0x20, 0x2f, 0x5c, 0x5f, 0x5f, 0x5f, 0x20, 0x5c, 0x20, 0xa, 0x2f, 0x5f, 0x5f, 0x5f, 0x5f, 0x5f, 0x20, 0x5c, 0x5c, 0x5f, 0x5f, 0x5f, 0x20, 0x20, 0x3e, 0x5f, 0x5f, 0x5f, 0x5f, 0x2f, 0x2f, 0x5f, 0x5f, 0x5f, 0x5f, 0x20, 0x20, 0x3e, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x5c, 0x2f, 0x20, 0x20, 0x20, 0x20, 0x5c, 0x2f, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x5c, 0x2f, 0x20, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x41, 0x20, 0x50, 0x6f, 0x77, 0x65, 0x72, 0x66, 0x75, 0x6c, 0x20, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x20, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d}), //++ TODO: optimize? (double allocation) or does compiler already optimize this?
	}
	file3 := &embedded.EmbeddedFile{
		Filename:    `bench.sh`,
		FileModTime: time.Unix(1486922764, 0),
		Content:     string([]byte{0x23, 0x21, 0x2f, 0x62, 0x69, 0x6e, 0x2f, 0x62, 0x61, 0x73, 0x68, 0xa, 0xa, 0x23, 0x20, 0x40, 0x7a, 0x65, 0x75, 0x73, 0x2d, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x3a, 0x20, 0x63, 0x6c, 0x65, 0x61, 0x6e, 0xa, 0x23, 0x20, 0x40, 0x7a, 0x65, 0x75, 0x73, 0x2d, 0x68, 0x65, 0x6c, 0x70, 0x3a, 0x20, 0x72, 0x75, 0x6e, 0x20, 0x74, 0x68, 0x65, 0x20, 0x62, 0x65, 0x6e, 0x63, 0x68, 0x6d, 0x61, 0x72, 0x6b, 0x73, 0xa, 0x23, 0x20, 0x40, 0x7a, 0x65, 0x75, 0x73, 0x2d, 0x61, 0x72, 0x67, 0x73, 0x3a, 0xa, 0xa}), //++ TODO: optimize? (double allocation) or does compiler already optimize this?
	}
	file4 := &embedded.EmbeddedFile{
		Filename:    `build.sh`,
		FileModTime: time.Unix(1486922764, 0),
		Content:     string([]byte{0x23, 0x21, 0x2f, 0x62, 0x69, 0x6e, 0x2f, 0x62, 0x61, 0x73, 0x68, 0xa, 0xa, 0x23, 0x20, 0x40, 0x7a, 0x65, 0x75, 0x73, 0x2d, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x3a, 0x20, 0x63, 0x6c, 0x65, 0x61, 0x6e, 0xa, 0x23, 0x20, 0x40, 0x7a, 0x65, 0x75, 0x73, 0x2d, 0x68, 0x65, 0x6c, 0x70, 0x3a, 0x20, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x20, 0x74, 0x68, 0x65, 0x20, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0xa, 0x23, 0x20, 0x40, 0x7a, 0x65, 0x75, 0x73, 0x2d, 0x61, 0x72, 0x67, 0x73, 0x3a, 0xa, 0xa}), //++ TODO: optimize? (double allocation) or does compiler already optimize this?
	}
	file5 := &embedded.EmbeddedFile{
		Filename:    `clean.sh`,
		FileModTime: time.Unix(1486922764, 0),
		Content:     string([]byte{0x23, 0x21, 0x2f, 0x62, 0x69, 0x6e, 0x2f, 0x62, 0x61, 0x73, 0x68, 0xa, 0xa, 0x23, 0x20, 0x40, 0x7a, 0x65, 0x75, 0x73, 0x2d, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x3a, 0x20, 0xa, 0x23, 0x20, 0x40, 0x7a, 0x65, 0x75, 0x73, 0x2d, 0x68, 0x65, 0x6c, 0x70, 0x3a, 0x20, 0x63, 0x6c, 0x65, 0x61, 0x6e, 0x20, 0x75, 0x70, 0x20, 0x74, 0x68, 0x65, 0x20, 0x6d, 0x65, 0x73, 0x73, 0xa, 0x23, 0x20, 0x40, 0x7a, 0x65, 0x75, 0x73, 0x2d, 0x61, 0x72, 0x67, 0x73, 0x3a, 0xa, 0xa}), //++ TODO: optimize? (double allocation) or does compiler already optimize this?
	}
	file6 := &embedded.EmbeddedFile{
		Filename:    `install.sh`,
		FileModTime: time.Unix(1486922764, 0),
		Content:     string([]byte{0x23, 0x21, 0x2f, 0x62, 0x69, 0x6e, 0x2f, 0x62, 0x61, 0x73, 0x68, 0xa, 0xa, 0x23, 0x20, 0x40, 0x7a, 0x65, 0x75, 0x73, 0x2d, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x3a, 0x20, 0x62, 0x75, 0x69, 0x6c, 0x64, 0xa, 0x23, 0x20, 0x40, 0x7a, 0x65, 0x75, 0x73, 0x2d, 0x68, 0x65, 0x6c, 0x70, 0x3a, 0x20, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x20, 0x61, 0x6e, 0x64, 0x20, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x20, 0x74, 0x6f, 0x20, 0x24, 0x50, 0x41, 0x54, 0x48, 0x20, 0xa, 0x23, 0x20, 0x40, 0x7a, 0x65, 0x75, 0x73, 0x2d, 0x61, 0x72, 0x67, 0x73, 0x3a, 0xa, 0xa}), //++ TODO: optimize? (double allocation) or does compiler already optimize this?
	}
	file7 := &embedded.EmbeddedFile{
		Filename:    `run.sh`,
		FileModTime: time.Unix(1486922764, 0),
		Content:     string([]byte{0x23, 0x21, 0x2f, 0x62, 0x69, 0x6e, 0x2f, 0x62, 0x61, 0x73, 0x68, 0xa, 0xa, 0x23, 0x20, 0x40, 0x7a, 0x65, 0x75, 0x73, 0x2d, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x3a, 0x20, 0x62, 0x75, 0x69, 0x6c, 0x64, 0xa, 0x23, 0x20, 0x40, 0x7a, 0x65, 0x75, 0x73, 0x2d, 0x68, 0x65, 0x6c, 0x70, 0x3a, 0x20, 0x72, 0x75, 0x6e, 0x20, 0x74, 0x68, 0x65, 0x20, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0xa, 0x23, 0x20, 0x40, 0x7a, 0x65, 0x75, 0x73, 0x2d, 0x61, 0x72, 0x67, 0x73, 0x3a, 0xa, 0xa}), //++ TODO: optimize? (double allocation) or does compiler already optimize this?
	}
	file8 := &embedded.EmbeddedFile{
		Filename:    `test.sh`,
		FileModTime: time.Unix(1486922764, 0),
		Content:     string([]byte{0x23, 0x21, 0x2f, 0x62, 0x69, 0x6e, 0x2f, 0x62, 0x61, 0x73, 0x68, 0xa, 0xa, 0x23, 0x20, 0x40, 0x7a, 0x65, 0x75, 0x73, 0x2d, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x3a, 0x20, 0x63, 0x6c, 0x65, 0x61, 0x6e, 0xa, 0x23, 0x20, 0x40, 0x7a, 0x65, 0x75, 0x73, 0x2d, 0x68, 0x65, 0x6c, 0x70, 0x3a, 0x20, 0x72, 0x75, 0x6e, 0x20, 0x74, 0x68, 0x65, 0x20, 0x74, 0x65, 0x73, 0x74, 0x73, 0xa, 0x23, 0x20, 0x40, 0x7a, 0x65, 0x75, 0x73, 0x2d, 0x61, 0x72, 0x67, 0x73, 0x3a, 0xa, 0xa}), //++ TODO: optimize? (double allocation) or does compiler already optimize this?
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   ``,
		DirModTime: time.Unix(1486922764, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file2, // ascii_art.txt
			file3, // bench.sh
			file4, // build.sh
			file5, // clean.sh
			file6, // install.sh
			file7, // run.sh
			file8, // test.sh

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`assets`, &embedded.EmbeddedBox{
		Name: `assets`,
		Time: time.Unix(1486922764, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"": dir1,
		},
		Files: map[string]*embedded.EmbeddedFile{
			"ascii_art.txt": file2,
			"bench.sh":      file3,
			"build.sh":      file4,
			"clean.sh":      file5,
			"install.sh":    file6,
			"run.sh":        file7,
			"test.sh":       file8,
		},
	})
}
