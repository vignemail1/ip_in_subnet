# ip_in_subnet

`ip_in_subnet` is a small, fast CLI tool written in Go that reads a CSV stream from **stdin** and prints every row whose subnet column (CIDR notation) **contains** a given IP address.

Both **IPv4** and **IPv6** are fully supported. It uses Go's modern `net/netip` package for correct, allocation-efficient prefix matching.

---

## Features

- Supports IPv4 and IPv6 CIDR notation (`192.0.2.0/24`, `2001:db8::/32`)
- Configurable CSV delimiter (`,` `;` `|` `\t` …)
- Optional header-line skip
- Configurable column index for the subnet field
- Exit code `0` if at least one row matches, `1` if none, `2` on usage/parse error
- Reads from stdin — composes naturally with pipes

---

## Installation

### From release binaries

Download the pre-built binary for your platform from the [Releases](https://github.com/vignemail1/ip_in_subnet/releases) page.

```bash
# Example for Linux amd64
curl -L https://github.com/vignemail1/ip_in_subnet/releases/latest/download/ip_in_subnet_linux_amd64.tar.gz | tar xz
sudo mv ip_in_subnet /usr/local/bin/
```

### From source

```bash
git clone https://github.com/vignemail1/ip_in_subnet.git
cd ip_in_subnet
go build -o ip_in_subnet .
```

Requires **Go 1.21+** (uses `net/netip`).

---

## Usage

```
Usage: ip_in_subnet [options] <ip>

Reads CSV from stdin and prints rows whose subnet column contains the given IP.

Options:
  -column int
        1-based column number containing the subnet CIDR (default 1)
  -delimiter string
        CSV field delimiter (single character, or \t for tab) (default ",")
  -skip-header
        skip the first line (header)
```

---

## Exit codes

| Code | Meaning |
|------|---------|
| `0`  | At least one row matched |
| `1`  | No row matched (or read error) |
| `2`  | Usage / argument error |

---

## Examples

### Basic — subnet in first column, no header

```csv
# subnets.csv
192.0.2.0/24,rouen,lan-a
10.0.0.0/8,paris,lan-b
2001:db8::/32,dc1,wan-v6
172.16.0.0/12,lyon,lan-c
```

```bash
cat subnets.csv | ip_in_subnet 192.0.2.10
# → 192.0.2.0/24,rouen,lan-a

cat subnets.csv | ip_in_subnet 10.42.0.1
# → 10.0.0.0/8,paris,lan-b

cat subnets.csv | ip_in_subnet 2001:db8::42
# → 2001:db8::/32,dc1,wan-v6
```

### With header line

```csv
# sites.csv
subnet,city,segment
192.0.2.0/24,rouen,lan-a
10.0.0.0/8,paris,lan-b
```

```bash
cat sites.csv | ip_in_subnet -skip-header 192.0.2.10
# → 192.0.2.0/24,rouen,lan-a
```

### Subnet in a different column

```csv
# infra.csv
site,description,subnet
rouen,office LAN,192.0.2.0/24
paris,datacenter,10.0.0.0/8
```

```bash
cat infra.csv | ip_in_subnet -skip-header -column 3 192.0.2.42
# → rouen,office LAN,192.0.2.0/24
```

### Semicolon-delimited file

```csv
# semicolon.csv
192.0.2.0/24;rouen;lan-a
10.0.0.0/8;paris;lan-b
```

```bash
cat semicolon.csv | ip_in_subnet -delimiter ';' 192.0.2.5
# → 192.0.2.0/24;rouen;lan-a
```

### Tab-separated file

```bash
cat subnets.tsv | ip_in_subnet -delimiter '\t' -skip-header -column 2 10.1.2.3
```

### Multiple matching subnets (overlapping prefixes)

If several rows match (e.g. both a `/8` supernet and a `/24`), **all matching rows** are printed:

```bash
cat subnets.csv | ip_in_subnet 10.0.1.1
# → 10.0.0.0/8,paris,lan-b
# → 10.0.0.0/24,paris,vlan-mgmt   ← also printed if present
```

### Use in a shell script

```bash
ip="192.0.2.55"
if cat subnets.csv | ip_in_subnet "$ip" > /dev/null 2>&1; then
    echo "$ip belongs to a known subnet"
else
    echo "$ip not found in any subnet"
fi
```

---

## Notes

- An IPv4 address will **not** match an IPv6-mapped prefix and vice versa (standard `net/netip` behaviour).
- Lines where the subnet column is empty or unparseable as CIDR are silently skipped.
- The output uses the **original delimiter** so the output is valid CSV/TSV for further piping.

---

## License

MIT
