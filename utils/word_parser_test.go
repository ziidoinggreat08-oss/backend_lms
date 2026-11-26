package utils

import (
	"bytes"
	"testing"
)

func TestParseDocxQuestions(t *testing.T) {
	docxBytes, err := GenerateSampleDocx()
	if err != nil {
		t.Fatalf("Gagal membuat contoh docx: %v", err)
	}

	reader := bytes.NewReader(docxBytes)
	soalList, err := ParseDocxQuestions(reader, int64(len(docxBytes)))
	if err != nil {
		t.Fatalf("Gagal mem-parsing docx: %v", err)
	}

	if len(soalList) != 3 {
		t.Fatalf("Ekspektasi 3 soal, tetapi didapat %d", len(soalList))
	}

	// Cek soal 1
	if soalList[0].Nomor != 1 {
		t.Errorf("Ekspektasi soal nomor 1, didapat %d", soalList[0].Nomor)
	}
	if len(soalList[0].Pilihan) != 4 {
		t.Errorf("Ekspektasi 4 pilihan di soal 1, didapat %d", len(soalList[0].Pilihan))
	}
	if soalList[0].Pilihan[0].Huruf != "A" || soalList[0].Pilihan[0].Teks != "TypeScript dan JavaScript" {
		t.Errorf("Pilihan A soal 1 tidak sesuai: %+v", soalList[0].Pilihan[0])
	}

	// Cek soal 2
	if soalList[1].Nomor != 2 {
		t.Errorf("Ekspektasi soal nomor 2, didapat %d", soalList[1].Nomor)
	}
	if len(soalList[1].Pilihan) != 4 {
		t.Errorf("Ekspektasi 4 pilihan di soal 2, didapat %d", len(soalList[1].Pilihan))
	}
	if soalList[1].Pilihan[1].Huruf != "B" || soalList[1].Pilihan[1].Teks != "MySQL" {
		t.Errorf("Pilihan B soal 2 tidak sesuai: %+v", soalList[1].Pilihan[1])
	}
}
