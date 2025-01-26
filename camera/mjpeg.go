package camera

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"

	"github.com/alexanderi96/libcamera-tray/utils"
)

// MJPEGFrameReader reads MJPEG frames from a stream
type MJPEGFrameReader struct {
	reader io.Reader
	buffer bytes.Buffer
}

// NewMJPEGFrameReader creates a new MJPEG frame reader
func NewMJPEGFrameReader(r io.Reader) *MJPEGFrameReader {
	return &MJPEGFrameReader{
		reader: r,
	}
}

// ReadFrame reads a single JPEG frame from the MJPEG stream
func (m *MJPEGFrameReader) ReadFrame() (image.Image, error) {
	// Look for JPEG start marker (0xFF 0xD8)
	startMarker := []byte{0xFF, 0xD8}
	endMarker := []byte{0xFF, 0xD9}

	// Use a larger buffer for more efficient reading
	buf := make([]byte, 65536) // Increased buffer size for larger frames
	
	for {
		n, err := m.reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				utils.Error("Error reading from stream: %v", err)
			}
			return nil, err
		}
		utils.Debug("Read %d bytes from stream", n)

		m.buffer.Write(buf[:n])
		data := m.buffer.Bytes()

		// Look for complete JPEG frame
		if len(data) > 4 {
			// Find start marker
			startIdx := -1
			for i := 0; i < len(data)-1; i++ {
				if bytes.Equal(data[i:i+2], startMarker) {
					startIdx = i
					break
				}
			}

			if startIdx >= 0 {
				// Find end marker after start marker
				for i := startIdx + 2; i < len(data)-1; i++ {
					if bytes.Equal(data[i:i+2], endMarker) {
						// Found complete frame
						frame := data[startIdx : i+2]
						
						// Decode frame
						img, err := jpeg.Decode(bytes.NewReader(frame))
						if err == nil {
							// Keep remaining data after frame
							m.buffer.Reset()
							if i+2 < len(data) {
								m.buffer.Write(data[i+2:])
							}
							return img, nil
						}
						utils.Error("Failed to decode JPEG frame: %v", err)
						
						// If decode failed, keep searching from next byte
						m.buffer.Reset()
						m.buffer.Write(data[startIdx+1:])
						break
					}
				}
				
				// If we found start but no end, keep from start
				if m.buffer.Len() == len(data) {
					m.buffer.Reset()
					m.buffer.Write(data[startIdx:])
				}
			} else {
				// No start marker found, keep last byte in case it's part of next frame
				m.buffer.Reset()
				m.buffer.WriteByte(data[len(data)-1])
			}
		}
	}
}
