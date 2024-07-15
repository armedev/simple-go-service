package main

import (
	"bufio"
	"bytes"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type CustomDb struct {
	path string
}

type Album struct {
	ID     string `json:"id,omitempty"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Price  int    `json:"price"`
}

type AlbumWithId struct {
	ID     string `json:"id"`
	Title  string `json:"title,omitempty"`
	Artist string `json:"artist,omitempty"`
	Price  *int   `json:"price,omitempty"`
}

func decodeAlbum(jobs <-chan string, results chan<- Album, wg *sync.WaitGroup) {
	defer wg.Done()

	for a := range jobs {
		data := strings.Split(a, ";")

		if len(data) != 4 {
			continue
		}

		price, err := strconv.Atoi(data[3])
		if err != nil {
			panic(err)
		}

		album := Album{
			ID:     data[0],
			Title:  data[1],
			Artist: data[2],
			Price:  price,
		}

		results <- album
	}
}

func encodeAlbum(jobs <-chan Album, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	for album := range jobs {
		data := strings.Join([]string{album.ID, album.Title, album.Artist, strconv.Itoa(int(album.Price))}, ";") + "\n"

		results <- data
	}
}

func (c *CustomDb) GetAlbums() ([]Album, error) {
	file, err := os.Open(c.path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	channelSize := 10

	jobs := make(chan string, channelSize)
	results := make(chan Album, channelSize)

	wg := new(sync.WaitGroup)

	for w := 1; w <= channelSize/2; w++ {
		wg.Add(1)
		go decodeAlbum(jobs, results, wg)
	}

	go func() {
		fileScanner := bufio.NewScanner(file)
		fileScanner.Split(bufio.ScanLines)
		for fileScanner.Scan() {
			jobs <- fileScanner.Text()
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	var albums []Album

	for v := range results {
		albums = append(albums, v)
	}

	return albums, nil
}

func (c *CustomDb) AddAlbums(albums []Album) ([]Album, error) {
	file, err := os.OpenFile(c.path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	channelSize := 10

	jobs := make(chan Album, channelSize)
	results := make(chan string, channelSize)

	updatedAlbums := make([]Album, 0)

	wg := new(sync.WaitGroup)

	for w := 1; w <= channelSize/2; w++ {
		wg.Add(1)
		go encodeAlbum(jobs, results, wg)
	}

	go func() {
		for _, album := range albums {
			id := album.ID
			if len(id) == 0 {
				id = uuid.NewString()
			}
			albumWithId := Album{
				ID:     id,
				Title:  album.Title,
				Artist: album.Artist,
				Price:  album.Price,
			}
			updatedAlbums = append(updatedAlbums, albumWithId)
			jobs <- albumWithId
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	var convertedAlbums []string

	for v := range results {
		convertedAlbums = append(convertedAlbums, v)
	}

	dataToAppend := strings.Join(convertedAlbums, "")

	if _, err := file.Write([]byte(dataToAppend)); err != nil {
		return nil, err
	}

	return updatedAlbums, nil
}

func (c *CustomDb) DeleteAlbums(keys []string) ([]string, error) {
	file, err := os.OpenFile(c.path, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	deletedKeys := make([]string, 0)

	fileScanner := bufio.NewScanner(file)

	var bts []byte

	buf := bytes.NewBuffer(bts)

	for fileScanner.Scan() {
		line := fileScanner.Text()
		id := strings.Split(line, ";")[0]

		if !slices.Contains(keys, id) {
			if _, err := buf.WriteString(line + "\n"); err != nil {
				return nil, err
			}
		} else {
			deletedKeys = append(deletedKeys, id)
		}
	}

	if len(deletedKeys) == 0 {
		return deletedKeys, nil
	}

	if err := file.Truncate(0); err != nil {
		return nil, err
	}
	if _, err := file.Seek(0, 0); err != nil {
		return nil, err
	}

	if _, err := buf.WriteTo(file); err != nil {
		return nil, err
	}

	return deletedKeys, nil
}

func getAlbum(albums []AlbumWithId, id string) *AlbumWithId {
	for _, album := range albums {
		if album.ID == id {
			return &album
		}
	}

	return nil
}

func (c *CustomDb) UpdateAlbums(albums []AlbumWithId) ([]Album, error) {
	file, err := os.OpenFile(c.path, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	updatedAlbums := make([]Album, 0)

	fileScanner := bufio.NewScanner(file)

	var bts []byte

	buf := bytes.NewBuffer(bts)

	linesChan := make(chan string, 1)

	albumsChan := make(chan Album, 1)

	go func() {
		for fileScanner.Scan() {
			line := fileScanner.Text()
			splittedLine := strings.Split(line, ";")
			id := splittedLine[0]

			album := getAlbum(albums, id)

			if album == nil {
				linesChan <- line
			} else {
				title := splittedLine[1]
				if len(album.Title) > 0 {
					title = album.Title
				}

				artist := splittedLine[2]
				if len(album.Artist) > 0 {
					artist = album.Artist
				}
				price := splittedLine[3]
				if album.Price != nil {
					price = strconv.Itoa(int(*album.Price))
				}

				linesChan <- strings.Join([]string{id, title, artist, price}, ";")

				intPrice, err := strconv.Atoi(price)
				if err != nil {
					panic(err)
				}

				albumsChan <- Album{
					ID:     id,
					Title:  title,
					Artist: artist,
					Price:  intPrice,
				}

			}
		}
		close(linesChan)
		close(albumsChan)
	}()

	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		defer wg.Done()
		for album := range albumsChan {
			updatedAlbums = append(updatedAlbums, album)
		}
	}()

	for line := range linesChan {
		if _, err := buf.WriteString(line + "\n"); err != nil {
			return nil, err
		}
	}

	wg.Wait()

	if len(updatedAlbums) == 0 {
		return updatedAlbums, nil
	}

	if err := file.Truncate(0); err != nil {
		return nil, err
	}
	if _, err := file.Seek(0, 0); err != nil {
		return nil, err
	}

	if _, err := buf.WriteTo(file); err != nil {
		return nil, err
	}

	return updatedAlbums, nil
}
