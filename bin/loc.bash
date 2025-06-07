echo "Lines of code per file type:"
for ext in go lox; do
  count=$(find . -name "*.${ext}" -not -path "./vendor/*" -not -path "./build/*" -exec cat {} + | wc -l)
  echo ".$ext: $count"
done

echo ""
total=$(find . -name "*.*" \( -name "*.go" -o -name "*.lox" \) -not -path "./vendor/*" -not -path "./build/*" -exec cat {} + | wc -l)
echo "Total LOC: $total"