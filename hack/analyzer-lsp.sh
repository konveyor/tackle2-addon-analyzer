usage() {
  echo "Usage: ${self} options"
  echo "  -h help" 
  echo "  --provider-settings"
  echo "  --output-file"
  echo "  --dep-output-file"
  echo "  --rules"
  echo "  --label-selector"
  echo "  --dep-label-selector"
}

RULES=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --provider-settings)
      SETTINGS="$2"
      shift 2
      ;;
    --provider-settings=*)
      SETTINGS="${1#*=}"
      shift
      ;;
    --output-file)
      OUTPUT="$2"
      shift 2
      ;;
    --output-file=*)
      OUTPUT="${1#*=}"
      shift
      ;;
    --dep-output)
      DEP_OUTPUT="$2"
      shift 2
      ;;
    --dep-output=*)
      DEP_OUTPUT="${1#*=}"
      shift
      ;;
    --rules)
      RULES+=("$2")
      shift 2
      ;;
    --rules=*)
      RULES+=("${1#*=}")
      shift
      ;;
    --label-selector)
      LABEL_SELECTOR="$2"
      shift 2
      ;;
    --label-selector=*)
      LABEL_SELECTOR="${1#*=}"
      shift
      ;;
    --dep-label-selector)
      DEP_LABEL_SELECTOR="$2"
      shift 2
      ;;
    --dep-label-selector=*)
      DEP_LABEL_SELECTOR="${1#*=}"
      shift
      ;;
    -h)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

echo
echo "Using:"
echo "  settings: ${SETTINGS}"
echo "     rules: ${RULES[*]}"
echo "    output: ${OUTPUT}"
echo "  selector: ${LABEL_SELECTOR}"
echo "  dep:"
echo "    output: ${DEP_OUTPUT}"
echo "  selector: ${DEP_LABEL_SELECTOR}"




