package file

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"maps"
	"os"
	"time"

	"github.com/fugo-app/fugo/internal/input"
	"github.com/fugo-app/fugo/pkg/debounce"
)

type fileWorker struct {
	path      string
	ext       map[string]string
	parser    fileParser
	rotator   fileRotator
	processor input.Processor

	offset   int64
	debounce *debounce.Debounce
}

func newFileWorker(
	path string,
	ext map[string]string,
	parser fileParser,
	rotator fileRotator,
	processor input.Processor,
) (*fileWorker, error) {
	return &fileWorker{
		path:      path,
		ext:       ext,
		parser:    parser,
		rotator:   rotator,
		processor: processor,
		offset:    getOffset(path),
		debounce:  nil,
	}, nil
}

func (fw *fileWorker) Start() {
	fw.debounce = debounce.NewDebounce(fw.tail, 250*time.Millisecond, true)
	fw.debounce.Start()
}

func (fw *fileWorker) Stop() {
	fw.debounce.Stop()
}

// Handle pushes the task to the debouncer
func (fw *fileWorker) Handle() {
	fw.debounce.Emit()
}

func (fw *fileWorker) tail() {
	file, err := os.Open(fw.path)
	if err != nil {
		return
	}
	defer file.Close()

	// Get file info to check size
	fileInfo, err := file.Stat()
	if err != nil {
		return
	}

	offset := fw.offset
	fileSize := fileInfo.Size()

	// If the file is empty, reset the offset to 0
	if fileSize == 0 {
		fw.offset = 0
		setOffset(fw.path, 0)
		return
	}

	// Check if file has been truncated (logrotate case)
	if offset > fileSize {
		offset = 0
	}

	_, err = file.Seek(offset, 0)
	if err != nil {
		return
	}

	reader := bufio.NewReaderSize(file, 64*1024)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			break
		}

		if !bytes.HasSuffix(line, []byte("\n")) {
			break
		}

		offset += int64(len(line))

		line = line[:len(line)-1]
		if bytes.HasSuffix(line, []byte("\r")) {
			line = line[:len(line)-1]
		}

		if len(line) > 0 {
			text := string(line)

			if raw, err := fw.parser.Parse(text); err == nil {
				maps.Copy(raw, fw.ext)
				if data := fw.processor.Serialize(raw); data != nil {
					fw.processor.Write(data)
				}
			}
		}

		if err == io.EOF {
			break
		}
	}

	// Update the offset for next run
	fw.offset = offset
	setOffset(fw.path, offset)

	if fw.rotator != nil {
		if fw.rotator.CheckSize(fileSize) {
			if err := fw.rotator.Rotate(fw.path); err != nil {
				log.Printf("failed to rotate log (%s): %v", fw.path, err)
				return
			}

			fw.offset = 0
			setOffset(fw.path, 0)
		}
	}
}
