package ipfs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
)

type Client struct {
	shell   *shell.Shell
	timeout time.Duration
}

func NewClient(apiURL string, timeout time.Duration) *Client {
	return &Client{
		shell:   shell.NewShell(apiURL),
		timeout: timeout,
	}
}

// Cat retrieves content from IPFS
func (c *Client) Cat(ctx context.Context, cid string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	reader, err := c.shell.Cat(cid)
	if err != nil {
		return nil, fmt.Errorf("failed to cat CID %s: %w", cid, err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	return data, nil
}

// Add adds content to IPFS
func (c *Client) Add(ctx context.Context, data []byte) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	reader := bytes.NewReader(data)
	cid, err := c.shell.Add(reader)
	if err != nil {
		return "", fmt.Errorf("failed to add content to IPFS: %w", err)
	}

	return cid, nil
}

// Pin pins content to local IPFS node
func (c *Client) Pin(ctx context.Context, cid string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	err := c.shell.Pin(cid)
	if err != nil {
		return fmt.Errorf("failed to pin CID %s: %w", cid, err)
	}

	return nil
}

// Unpin unpins content from local IPFS node
func (c *Client) Unpin(ctx context.Context, cid string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	err := c.shell.Unpin(cid)
	if err != nil {
		return fmt.Errorf("failed to unpin CID %s: %w", cid, err)
	}

	return nil
}

// Exists checks if content exists on IPFS network
func (c *Client) Exists(ctx context.Context, cid string) bool {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Try to get object stats to check existence
	_, err := c.shell.ObjectStat(cid)
	return err == nil
}

// GetSize gets the size of content
func (c *Client) GetSize(ctx context.Context, cid string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	stat, err := c.shell.ObjectStat(cid)
	if err != nil {
		return 0, fmt.Errorf("failed to get object stats for CID %s: %w", cid, err)
	}

	return int64(stat.CumulativeSize), nil
}

// ListPins lists pinned content
func (c *Client) ListPins(ctx context.Context) (map[string]shell.PinInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	pins, err := c.shell.Pins()
	if err != nil {
		return nil, fmt.Errorf("failed to list pins: %w", err)
	}

	return pins, nil
}
