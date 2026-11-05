package domain

type Paragraph struct {
	ID   string
	Text string
}

type Document struct {
	ID         string
	Name       string
	Kind       string
	ContractID string
	Status     string
	Paragraphs []Paragraph
}

type Evidence struct {
	DocumentID  string
	FileName    string
	ParagraphID string
	Quote       string
}

type Field struct {
	Value    string
	Evidence *Evidence
	Changed  bool
}

type Notice struct {
	Kind     string
	Message  string
	Evidence Evidence
}

type Contract struct {
	ID       string
	Supplier Field
	Amount   Field
	Deadline Field
	Notices  []Notice
	Review   bool
}

type Result struct {
	AsOf      string
	Contracts []Contract
	Documents []Document
	Warnings  []string
}
