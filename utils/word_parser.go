package utils

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type ParsedPilihan struct {
	Huruf   string `json:"huruf"`
	Teks    string `json:"teks"`
	IsBenar bool   `json:"is_benar"`
}

type ParsedSoal struct {
	Nomor      int             `json:"nomor"`
	Pertanyaan string          `json:"pertanyaan"`
	Bobot      int             `json:"bobot"`
	Pilihan    []ParsedPilihan `json:"pilihan"`
}

var (
	reQuestion = regexp.MustCompile(`^(?i)(?:soal\s*)?(\d+)[\.\)\:\-]\s*(.*)$`)
	reChoice   = regexp.MustCompile(`^([A-Ea-e])[\.\)\:\-]\s*(.*)$`)
	reKey      = regexp.MustCompile(`^(?i)(?:kunci(?:\s*jawaban)?|jawaban)\s*[\:\=]\s*([A-Ea-e])\b`)
)

// ExtractTextFromDocx membaca seluruh paragraf dari file Word (.docx)
func ExtractTextFromDocx(r io.ReaderAt, size int64) ([]string, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("berkas bukan format docx (ZIP) yang valid: %w", err)
	}

	var docFile *zip.File
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			docFile = f
			break
		}
	}

	if docFile == nil {
		return nil, errors.New("tidak dapat menemukan word/document.xml di dalam berkas docx")
	}

	rc, err := docFile.Open()
	if err != nil {
		return nil, fmt.Errorf("gagal membuka isi dokumen docx: %w", err)
	}
	defer rc.Close()

	decoder := xml.NewDecoder(rc)
	var paragraphs []string
	var currentParagraph strings.Builder
	inParagraph := false
	inText := false

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("gagal membaca struktur XML docx: %w", err)
		}

		switch elem := token.(type) {
		case xml.StartElement:
			if elem.Name.Local == "p" {
				inParagraph = true
				currentParagraph.Reset()
			} else if elem.Name.Local == "t" {
				inText = true
			}
		case xml.EndElement:
			if elem.Name.Local == "p" {
				inParagraph = false
				line := strings.TrimSpace(currentParagraph.String())
				// Ganti non-breaking space (0xA0) dengan spasi biasa
				line = strings.ReplaceAll(line, "\u00a0", " ")
				line = strings.TrimSpace(line)
				if line != "" {
					paragraphs = append(paragraphs, line)
				}
			} else if elem.Name.Local == "t" {
				inText = false
			}
		case xml.CharData:
			if inParagraph && inText {
				currentParagraph.Write(elem)
			}
		}
	}

	return paragraphs, nil
}

// ParseDocxQuestions mengubah paragraf-paragraf dari docx menjadi struktur soal & opsi pilihan ganda
func ParseDocxQuestions(r io.ReaderAt, size int64) ([]ParsedSoal, error) {
	paragraphs, err := ExtractTextFromDocx(r, size)
	if err != nil {
		return nil, err
	}

	return ParseParagraphsToQuestions(paragraphs)
}

// ParseParagraphsToQuestions mengelompokkan baris teks menjadi soal dan pilihan ganda
func ParseParagraphsToQuestions(paragraphs []string) ([]ParsedSoal, error) {
	var soalList []ParsedSoal
	var currentSoal *ParsedSoal

	for _, rawLine := range paragraphs {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		// 1. Cek apakah baris ini adalah nomor soal baru (Contoh: "1. Apa itu HTML?" atau "Soal 1: ...")
		if match := reQuestion.FindStringSubmatch(line); match != nil {
			num, _ := strconv.Atoi(match[1])
			if currentSoal != nil && len(currentSoal.Pilihan) > 0 {
				soalList = append(soalList, *currentSoal)
			}
			currentSoal = &ParsedSoal{
				Nomor:      num,
				Pertanyaan: strings.TrimSpace(match[2]),
				Bobot:      1,
				Pilihan:    []ParsedPilihan{},
			}
			continue
		}

		// 2. Cek apakah baris ini adalah pilihan jawaban (Contoh: "A. HyperText Markup Language" atau "b) ...")
		if match := reChoice.FindStringSubmatch(line); match != nil && currentSoal != nil {
			huruf := strings.ToUpper(match[1])
			teks := strings.TrimSpace(match[2])
			currentSoal.Pilihan = append(currentSoal.Pilihan, ParsedPilihan{
				Huruf:   huruf,
				Teks:    teks,
				IsBenar: false,
			})
			continue
		}

		// 3. Cek apakah ada baris penanda kunci jawaban (Contoh: "Kunci: A" atau "Kunci Jawaban: B")
		if match := reKey.FindStringSubmatch(line); match != nil && currentSoal != nil {
			kunci := strings.ToUpper(match[1])
			for i := range currentSoal.Pilihan {
				if currentSoal.Pilihan[i].Huruf == kunci {
					currentSoal.Pilihan[i].IsBenar = true
				}
			}
			continue
		}

		// 4. Jika bukan nomor soal baru, bukan opsi, dan bukan kunci:
		// Jika opsi belum dibuat, tambahkan ke teks pertanyaan (mendukung pertanyaan multi-baris)
		if currentSoal != nil {
			if len(currentSoal.Pilihan) == 0 {
				if currentSoal.Pertanyaan != "" {
					currentSoal.Pertanyaan += "\n" + line
				} else {
					currentSoal.Pertanyaan = line
				}
			} else {
				// Jika opsi sudah ada, tambahkan ke teks opsi terakhir
				lastIdx := len(currentSoal.Pilihan) - 1
				currentSoal.Pilihan[lastIdx].Teks += " " + line
			}
		}
	}

	if currentSoal != nil && len(currentSoal.Pilihan) > 0 {
		soalList = append(soalList, *currentSoal)
	}

	// Rapikan penomoran jika berantakan
	for i := range soalList {
		soalList[i].Nomor = i + 1
	}

	if len(soalList) == 0 {
		return nil, errors.New("tidak ada format soal yang dapat dikenali. Pastikan soal memiliki nomor (1., 2.) dan pilihan (A., B., C., D.)")
	}

	return soalList, nil
}

// GenerateSampleDocx membuat file docx contoh format soal yang bisa diunduh guru
func GenerateSampleDocx() ([]byte, error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	// [Content_Types].xml
	contentTypes := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`
	if w, err := zw.Create("[Content_Types].xml"); err != nil {
		return nil, err
	} else {
		w.Write([]byte(contentTypes))
	}

	// _rels/.rels
	rels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`
	if w, err := zw.Create("_rels/.rels"); err != nil {
		return nil, err
	} else {
		w.Write([]byte(rels))
	}

	// word/document.xml with 3 sample questions
	sampleXml := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>1. Bahasa pemrograman apa yang digunakan untuk membangun antarmuka web modern?</w:t></w:r></w:p>
    <w:p><w:r><w:t>A. TypeScript dan JavaScript</w:t></w:r></w:p>
    <w:p><w:r><w:t>B. Assembly</w:t></w:r></w:p>
    <w:p><w:r><w:t>C. Pascal</w:t></w:r></w:p>
    <w:p><w:r><w:t>D. COBOL</w:t></w:r></w:p>
    <w:p><w:r><w:t>2. Manakah di bawah ini yang merupakan sistem manajemen basis data relasional (RDBMS)?</w:t></w:r></w:p>
    <w:p><w:r><w:t>A. Redis</w:t></w:r></w:p>
    <w:p><w:r><w:t>B. MySQL</w:t></w:r></w:p>
    <w:p><w:r><w:t>C. MongoDB</w:t></w:r></w:p>
    <w:p><w:r><w:t>D. Cassandra</w:t></w:r></w:p>
    <w:p><w:r><w:t>3. Protokol jaringan apa yang standar digunakan untuk mengamankan komunikasi web (enkripsi)?</w:t></w:r></w:p>
    <w:p><w:r><w:t>A. HTTP</w:t></w:r></w:p>
    <w:p><w:r><w:t>B. FTP</w:t></w:r></w:p>
    <w:p><w:r><w:t>C. HTTPS</w:t></w:r></w:p>
    <w:p><w:r><w:t>D. Telnet</w:t></w:r></w:p>
  </w:body>
</w:document>`
	if w, err := zw.Create("word/document.xml"); err != nil {
		return nil, err
	} else {
		w.Write([]byte(sampleXml))
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
