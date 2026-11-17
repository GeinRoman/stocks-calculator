export function observeSidebarViewport(media: MediaQueryList, onChange: (narrow: boolean) => void): () => void {
	const update = () => onChange(media.matches);
	update();
	media.addEventListener('change', update);
	return () => media.removeEventListener('change', update);
}
