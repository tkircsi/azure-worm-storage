package handler

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/gin-gonic/gin"
)

type Item struct {
	FULKID string `json:"fulkid"`
	Time string `json:"time"`
	SHA256 string `json:"sha256"`
	// Payload  string `json:"payload"`
}

type RequestItem struct {
	FULKID string `json:"fulkid"`
	Time string `json:"time"`
	// Payload  string `json:"payload"`
}

var p pipeline.Pipeline
var baseURL string

func init() {
	accName := os.Getenv("AZURE_STORAGE_ACCOUNT")
	accKey := os.Getenv("AZURE_STORAGE_ACCESS_KEY")
	accContainer := os.Getenv("AZURE_STORAGE_CONTAINER")
	cred, err := azblob.NewSharedKeyCredential(accName, accKey)
	if err != nil {
		log.Fatal("Invalid credentials with error: " + err.Error())
	}

	p = azblob.NewPipeline(cred, azblob.PipelineOptions{})
	baseURL = fmt.Sprintf("https://%s.blob.core.windows.net/%s", accName, accContainer)
	fmt.Printf("Connected to azure storage: %s\n", accName)
}

func Add() gin.HandlerFunc {
	return func (c *gin.Context) {
		var item RequestItem

		if c.ShouldBind(&item) != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid item data.",
			})
			return
		}

		ctx := context.Background()
		fname := fmt.Sprintf("%s_%s", item.FULKID, item.Time)
		h := sha256.New()
		h.Write([]byte(fname))
		fname = fmt.Sprintf("%s_%x", fname, h.Sum(nil))

		URL, _ := url.Parse(baseURL)
		containerURL := azblob.NewContainerURL(*URL, p)

		blobURL := containerURL.NewBlockBlobURL(fname)
		data := []byte("")
		_, err := azblob.UploadBufferToBlockBlob(ctx, data, blobURL, azblob.UploadToBlockBlobOptions{} )
		if err != nil {
			if serr, ok := err.(azblob.StorageError); ok { // This error is a Service-specific
				c.JSON(http.StatusBadRequest, gin.H{
					"error": serr.ServiceCode(),
				})
				return
			}
		}

		c.JSON(http.StatusCreated, gin.H{
			"claim": fname,
		})
	}
}

func GetByPrefix() gin.HandlerFunc {
	return func (c *gin.Context)  {
		prefix := c.Query("prefix")
		var items []Item
		ctx := context.Background()
		// u := fmt.Sprintf(baseURL)
		listURL, _ := url.Parse(baseURL)
		containerURL := azblob.NewContainerURL(*listURL, p)

		// for marker := (azblob.Marker{}); marker.NotDone(); {
			marker := azblob.Marker{}
			// Get a result segment starting with the blob indicated by the current Marker.
			listBlob, _ := containerURL.ListBlobsFlatSegment(ctx, marker, azblob.ListBlobsSegmentOptions{
				Prefix: prefix,
				MaxResults: 10,
			})

			// ListBlobs returns the start of the next segment; you MUST use this to get
			// the next segment (after processing the current result segment).
			// marker = listBlob.NextMarker

			// Process the blobs returned in this result segment (if the segment is empty, the loop body won't execute)
			for _, blobInfo := range listBlob.Segment.BlobItems {
				infos := strings.Split(blobInfo.Name, "_")
				item := Item{
					FULKID: infos[0],
					Time: infos[1],
					SHA256: infos[2],
				}
				items = append(items, item)
			}		
//			}
		c.JSON(http.StatusOK, items)
	}
}
