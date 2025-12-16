package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// ==========================================
// 1. KONFIGURASI PROTOKOL (PHYSICS ENGINE)
// ==========================================
const (
	TargetBlockTime    = 60   // Target 1 menit per blok
	MaxPropagationTime = 5    // Batas waktu toleransi jaringan (5 detik)
	UtilizationRatio   = 0.08 // 8% Bandwidth terpakai (Kalibrasi agar lolos validasi)
)

// ==========================================
// 2. STRUKTUR DATA (CORE ARCHITECTURE)
// ==========================================

// Transaction: Unit terkecil data
type Transaction struct {
	ID        string
	Sender    string
	Receiver  string
	Amount    float64
	Timestamp int64
}

// Block: Kontainer data
type Block struct {
	Index        int
	Timestamp    int64
	PrevHash     string
	Transactions []Transaction
	MerkleRoot   string  // Kunci efisiensi L2
	SizeKB       float64 // Ukuran dinamis
	MinerID      string
	Nonce        int
	Hash         string
}

// ==========================================
// 3. FUNGSI KRIPTOGRAFI (SECURITY)
// ==========================================

// Membuat Hash SHA-256
func calculateHash(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// Merkle Tree: Mengubah ribuan transaksi menjadi 1 Hash Root
func BuildMerkleRoot(txs []Transaction) string {
	if len(txs) == 0 {
		return ""
	}
	var hashes []string
	for _, tx := range txs {
		hashes = append(hashes, calculateHash(tx.ID+tx.Sender+tx.Receiver))
	}

	// Loop sampai sisa 1 hash (Root)
	for len(hashes) > 1 {
		var newLevel []string
		for i := 0; i < len(hashes); i += 2 {
			leaf1 := hashes[i]
			leaf2 := leaf1
			if i+1 < len(hashes) {
				leaf2 = hashes[i+1]
			}
			newLevel = append(newLevel, calculateHash(leaf1+leaf2))
		}
		hashes = newLevel
	}
	return hashes[0]
}

// ==========================================
// 4. LOGIKA JARINGAN (BANDWIDTH VALIDATION)
// ==========================================

// Rumus: MaxBlockSize = (Speed * Ratio * Time)
func CalculateMaxBlockSize(speedMbps float64) float64 {
	// Hasil dalam KB
	return (speedMbps * UtilizationRatio * float64(TargetBlockTime) / 8) * 1024
}

// Polisi Jaringan: Menolak blok yang "berat" di jalan
func SimulatePropagation(block Block, realSpeedMbps float64) bool {
	// Fisika: Waktu = Ukuran / Kecepatan
	sizeMb := (block.SizeKB / 1024) * 8
	actualTime := sizeMb / realSpeedMbps

	fmt.Printf("\n[Network Validation]\n")
	fmt.Printf("   📡 Blok Size  : %.2f KB\n", block.SizeKB)
	fmt.Printf("   🚀 Speed Miner: %.0f Mbps\n", realSpeedMbps)
	fmt.Printf("   ⏱️  Waktu Kirim: %.4f detik (Max: %d detik)\n", actualTime, MaxPropagationTime)

	if actualTime > MaxPropagationTime {
		fmt.Printf("   ❌ REJECTED: Blok terlalu besar, jaringan macet!\n")
		return false
	}
	fmt.Printf("   ✅ ACCEPTED: Blok valid & terpropagasi.\n")
	return true
}

// ==========================================
// 5. ENGINE MINING (WORKER)
// ==========================================

func MineBlock(prevBlock Block, mempool []Transaction, minerName string, speedMbps float64) Block {
	fmt.Printf("\n==================================================\n")
	fmt.Printf("⛏️  MINING PROCESS: %s (Speed: %.0f Mbps)\n", minerName, speedMbps)
	fmt.Printf("==================================================\n")

	// 1. Hitung Batas Ukuran Blok (Dynamic)
	maxSizeKB := CalculateMaxBlockSize(speedMbps)
	fmt.Printf(">> Kapasitas Blok Max: %.2f KB\n", maxSizeKB)

	// 2. Pilih Transaksi (Greedy)
	var selectedTxs []Transaction
	currentSize := 0.0
	txSizeAvg := 0.5 // Asumsi 1 tx = 0.5 KB

	for _, tx := range mempool {
		if currentSize+txSizeAvg > maxSizeKB {
			break // Berhenti jika penuh
		}
		selectedTxs = append(selectedTxs, tx)
		currentSize += txSizeAvg
	}
	fmt.Printf(">> Mengangkut %d Transaksi dari Mempool.\n", len(selectedTxs))

	// 3. Bangun Merkle Tree
	merkleRoot := BuildMerkleRoot(selectedTxs)
	fmt.Printf(">> Merkle Root: %s...\n", merkleRoot[:15])

	// 4. Assembly Blok
	newBlock := Block{
		Index:        prevBlock.Index + 1,
		Timestamp:    time.Now().Unix(),
		PrevHash:     prevBlock.PrevHash, // Simplified chain linking
		Transactions: selectedTxs,
		MerkleRoot:   merkleRoot,
		SizeKB:       currentSize,
		MinerID:      minerName,
		Nonce:        0,
	}

	// 5. Proof of Work (Simulasi Cepat)
	// Kita cari hash yang depannya "00"
	fmt.Print(">> Hashing: ")
	for {
		header := fmt.Sprintf("%d%s%s%d", newBlock.Index, newBlock.MerkleRoot, newBlock.PrevHash, newBlock.Nonce)
		newBlock.Hash = calculateHash(header)
		
		if newBlock.Hash[:2] == "00" { // Difficulty rendah untuk demo
			fmt.Printf("FOUND! Nonce: %d\n", newBlock.Nonce)
			fmt.Printf(">> Block Hash: %s...\n", newBlock.Hash[:20])
			break
		}
		newBlock.Nonce++
		if newBlock.Nonce%50000 == 0 {
			fmt.Print(".")
		}
	}

	return newBlock
}

// ==========================================
// 6. L1 CHECKPOINT (ROOTING TO BITCOIN)
// ==========================================
func GenerateCheckpoint(b Block) {
	// Format: RNR + Index + MerkleRoot
	prefix := "524e52" // Hex untuk "RNR"
	indexHex := fmt.Sprintf("%08x", b.Index)
	payload := prefix + indexHex + b.MerkleRoot

	fmt.Printf("\n🔒 [L1 CHECKPOINT] Mengirim ke Bitcoin...\n")
	fmt.Printf("   Payload OP_RETURN : %s\n", payload)
	fmt.Printf("   Size              : %d bytes (Hemat biaya!)\n", len(payload)/2)
}

// ==========================================
// MAIN PROGRAM (SIMULATION LOOP)
// ==========================================
func main() {
	rand.Seed(time.Now().UnixNano())

	// 1. Inisialisasi Genesis
	chain := []Block{{Index: 0, Hash: "0000Genesis", MerkleRoot: "000", SizeKB: 0}}
	
	fmt.Println("🚀 ROUTE N ROOT (RnR) - FULL NODE SIMULATION")
	fmt.Println("Status: ONLINE")
	
	// 2. Simulasi Mempool (User Spam)
	// Kita buat 20.000 transaksi dummy
	var mempool []Transaction
	fmt.Print("\n[System] Menerima transaksi masuk...")
	for i := 0; i < 20000; i++ {
		mempool = append(mempool, Transaction{
			ID:     strconv.Itoa(i),
			Sender: fmt.Sprintf("Wallet_%d", rand.Intn(1000)),
			Amount: rand.Float64() * 10,
		})
	}
	fmt.Printf(" SELESAI. Total Mempool: %d Tx.\n", len(mempool))

	// 3. MINING LOOP (Simulasi 3 Blok)
	// Kita simulasikan 3 blok berurutan
	for i := 1; i <= 3; i++ {
		prevBlock := chain[len(chain)-1]
		
		// Variasi Kecepatan Miner (Supaya dinamis)
		// Blok 1: 50 Mbps, Blok 2: 100 Mbps, Blok 3: 20 Mbps
		minerSpeed := float64(50)
		if i == 2 { minerSpeed = 100 }
		if i == 3 { minerSpeed = 20 }

		minerName := fmt.Sprintf("Miner_Node_%d", i)

		// A. Mining
		newBlock := MineBlock(prevBlock, mempool, minerName, minerSpeed)
		
		// B. Validasi Jaringan
		if SimulatePropagation(newBlock, minerSpeed) {
			// C. Tambah ke Chain
			chain = append(chain, newBlock)
			
			// D. Rooting ke Bitcoin
			GenerateCheckpoint(newBlock)
			
			// Hapus transaksi yang sudah diproses dari mempool (Sederhana)
			processedCount := len(newBlock.Transactions)
			if processedCount < len(mempool) {
				mempool = mempool[processedCount:]
			}
		} else {
			fmt.Println("❌ BLOK DIBUANG (ORPHANED)")
		}
		
		time.Sleep(1 * time.Second) // Jeda visual
	}

	fmt.Printf("\n✅ SIMULASI SELESAI. Panjang Chain: %d Blok.\n", len(chain))
}
