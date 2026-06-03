package containercreate

import (
	"crypto/rand"
	"math/big"
	"time"

	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
	Str "github.com/IBM/fp-go/v2/string"
)

// adjectives for container names (Docker-style naming).
var adjectives = []string{
	"admiring", "adoring", "affectionate", "agitated", "amazing",
	"angry", "awesome", "beautiful", "blissful", "bold", "boring",
	"brave", "busy", "calm", "charming", "clever", "cool", "compassionate",
	"competent", "condescending", "confident", "cranky", "crazy", "curious",
	"dazzling", "determined", "distracted", "dreamy", "eager", "ecstatic",
	"elastic", "elated", "elegant", "eloquent", "epic", "exciting", "fervent",
	"festive", "flamboyant", "focused", "friendly", "frosty", "funny", "gallant",
	"gifted", "goofy", "gracious", "great", "groggy", "happy", "hardcore",
	"heuristic", "hopeful", "hungry", "infallible", "inspiring", "intelligent",
	"interesting", "introspective", "jolly", "jovial", "keen", "kind", "laughing",
	"loving", "lucid", "magical", "mystifying", "modest", "musing", "naughty",
	"nervous", "nice", "nifty", "nostalgic", "objective", "optimistic", "peaceful",
	"pedantic", "pensive", "practical", "priceless", "quirky", "quizzical",
	"recursing", "relaxed", "reverent", "romantic", "sad", "serene", "sharp",
	"silly", "sleepy", "stoic", "strange", "stupefied", "suspicious", "sweet",
	"tender", "thirsty", "trusting", "unruffled", "upbeat", "vibrant", "vigilant",
	"vigorous", "wizardly", "wonderful", "xenodochial", "youthful", "zealous", "zen",
}

// nouns for container names (famous scientists and engineers).
var nouns = []string{
	"albattani", "allen", "almeida", "antonelli", "archimedes", "ardinghelli",
	"aryabhata", "austin", "babbage", "banach", "bardeen", "bartik", "bassi",
	"beaver", "bell", "benz", "black", "blackwell", "bohr", "booth", "borg",
	"bose", "bouman", "boyd", "brahmagupta", "brattain", "brown", "buck",
	"burnell", "cannon", "carson", "cartwright", "carver", "cauchy", "cerf",
	"chandrasekhar", "chaplygin", "chatelet", "chatterjee", "chebyshev", "cohen",
	"chaum", "clarke", "colden", "cori", "cray", "curie", "curran", "darwin",
	"davinci", "dewdney", "dhawan", "diffie", "dijkstra", "dirac", "driscoll",
	"dubinsky", "easley", "edison", "einstein", "elbakyan", "elgamal", "elion",
	"ellis", "engelbart", "euclid", "euler", "faraday", "feistel", "fermat",
	"fermi", "feynman", "franklin", "gagarin", "galileo", "galois", "ganguly",
	"gates", "gauss", "germain", "goldberg", "goldstine", "goldwasser", "golick",
	"goodall", "gould", "greider", "grothendieck", "haibt", "hall", "hamilton",
	"haslett", "hawking", "hellman", "heisenberg", "hermann", "herschel", "hertz",
	"heyrovsky", "hodgkin", "hoover", "hopper", "hugle", "hypatia", "ishizaka",
	"jackson", "jang", "jemison", "jennings", "jepsen", "johnson", "joliot",
	"jones", "kalam", "kapitsa", "kare", "keldysh", "keller", "kepler", "khayyam",
	"khorana", "kilby", "kirch", "knuth", "kowalevski", "lalande", "lamarr",
	"lamport", "leakey", "leavitt", "lederberg", "lehmann", "lewin", "lichterman",
	"liskov", "lovelace", "lumiere", "mahavira", "margulis", "matsumoto", "maxwell",
	"mayer", "mccarthy", "mcclintock", "mclaurin", "mclean", "mcnulty", "mendel",
	"mendeleev", "meitner", "meninsky", "merkle", "mestorf", "mirzakhani", "montalcini",
	"moore", "morse", "murdock", "moser", "napier", "nash", "neumann", "newton",
	"nightingale", "nobel", "noether", "northcutt", "noyce", "panini", "pare",
	"pascal", "pasteur", "payne", "perlman", "pike", "poincare", "poitras",
	"proskuriakova", "ptolemy", "raman", "ramanujan", "ride", "ritchie", "rhodes",
	"robinson", "roentgen", "rosalind", "rubin", "saha", "sammet", "sanderson",
	"satoshi", "shamir", "shannon", "shaw", "shirley", "shockley", "shtern", "sinoussi",
	"snyder", "solomon", "spence", "stonebraker", "sutherland", "swanson", "swartz",
	"swirles", "taussig", "tereshkova", "tesla", "tharp", "thompson", "torvalds",
	"tu", "turing", "tyson", "varahamihira", "vaughan", "vaughn", "villani",
	"visvesvaraya", "volhard", "wescoff", "wilbur", "wiles", "williams", "williamson",
	"wilson", "wing", "wozniak", "wright", "wu", "yalow", "yonath", "zhukovsky",
}

// GenerateName creates a random container name in the format "adjective-noun".
func GenerateName() string {
	return Str.IntersperseSemigroup("-").Concat(
		pickRandom(adjectives),
		pickRandom(nouns),
	)
}

// pickRandom selects a random element from a slice using fp-go.
func pickRandom[T any](items []T) T {
	return F.Pipe2(
		items,
		A.Lookup[T](randomInt(len(items))),
		O.GetOrElse(func() T { var zero T; return zero }),
	)
}

// randomInt returns a cryptographically secure random integer in [0, max).
func randomInt(limit int) int {
	return F.Pipe1(
		O.FromPredicate(func(int) bool { return limit > 0 })(limit),
		O.Fold(
			func() int { return 0 },
			func(m int) int {
				n, err := rand.Int(rand.Reader, big.NewInt(int64(m)))

				return F.Pipe1(
					O.FromPredicate(func(error) bool { return err == nil })(err),
					O.Fold(
						func() int {
							return int(time.Now().UnixNano()) % m
						},
						func(error) int { return int(n.Int64()) },
					),
				)
			},
		),
	)
}
