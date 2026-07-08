export function useGlobalSearchState() {
  const open = useState('global-search-open', () => false)
  return open
}
