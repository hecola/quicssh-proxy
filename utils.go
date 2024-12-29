package main

import (
	"context"
	"io"
)

func ReadAndWrite(ctx context.Context, src io.Reader, dst io.Writer) error {
	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			n, err := src.Read(buf)
			if n > 0 {
				if _, writeErr := dst.Write(buf[:n]); writeErr != nil {
					return writeErr
				}
			}
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}
	}
}
