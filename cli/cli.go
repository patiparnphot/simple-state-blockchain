package cli

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/patiparnphot/simple-state-blockchain/blockchain"
	"github.com/patiparnphot/simple-state-blockchain/merkletrie"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" getbalance -address ADDRESS - get the balance for an address")
	fmt.Println(" createblockchain -address ADDRESS creates a blockchain and sends genesis reward to address")
	fmt.Println(" printchain - Prints the blocks in the chain")
	fmt.Println(" send -from FROM -to TO -amount AMOUNT - Send amount of coins")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) printChain() {
	chain, _ := blockchain.ResumeBlockChain()
	defer chain.Database.Close()

	iter := chain.Iterator()

	for {
		block := iter.Next()

		fmt.Printf("Previous hash: %x\n", block.PrevHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Nonce: %d\n", block.Nonce)

		pow := blockchain.NewProof(block)

		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Printf("Transaction Inputs: %v\n", tx.Inputs)
			fmt.Printf("Transaction Outputs: %v\n", tx.Outputs)
		}
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) createBlockchain(address string) {
	chain, _ := blockchain.InitBlockChain(address)
	chain.Database.Close()
	fmt.Println("Finished!!!")
}

func (cli *CommandLine) getBalance(address string) {
	chain, accList := blockchain.ResumeBlockChain()
	defer chain.Database.Close()

	balance := 0
	// UTXOs := chain.FindUTXO(address)

	// for _, out := range UTXOs {
	// 	balance += out.Value
	// }

	trie := merkletrie.NewTrie(accList)

	IsVertify := trie.VertifyAccount(merkletrie.Account{Address: address})

	if IsVertify {
		for _, acc := range accList {
			if acc.Address == address {
				balance = acc.Balance
			}
		}
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int) {
	chain, accList := blockchain.ResumeBlockChain()
	defer chain.Database.Close()

	trie := merkletrie.NewTrie(accList)

	tx := blockchain.NewTransaction(from, to, amount, chain)

	IsFromVertify := trie.VertifyAccount(merkletrie.Account{Address: from})
	if IsFromVertify {
		for _, acc := range accList {
			if acc.Address == from {
				acc.Balance -= amount
			}
		}
	}

	IsToVertify := trie.VertifyAccount(merkletrie.Account{Address: to})
	if IsToVertify {
		for _, acc := range accList {
			if acc.Address == to {
				acc.Balance += amount
			}
		}
	} else {
		accList = append(accList, &merkletrie.Account{
			Address: to,
			Balance: amount,
		})
	}

	chain.AddBlock([]*blockchain.Transaction{tx}, accList, trie)

	fmt.Println("Transfer Success!!!")
}

func (cli *CommandLine) Run() {
	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "the address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "the address to send genesis reward to")
	sendFrom := sendCmd.String("from", "", "sender address")
	sendTo := sendCmd.String("to", "", "receiver address")
	sendAmount := sendCmd.Int("amount", 0, "amount to send")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "send":
		err := sendCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		} else {
			cli.getBalance(*getBalanceAddress)
		}
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		} else {
			cli.createBlockchain(*createBlockchainAddress)
		}
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		} else {
			cli.send(*sendFrom, *sendTo, *sendAmount)
		}
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}
