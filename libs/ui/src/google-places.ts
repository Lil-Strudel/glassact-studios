let placesLibraryPromise: Promise<google.maps.PlacesLibrary> | null = null;

export function loadGooglePlaces(
  apiKey: string,
): Promise<google.maps.PlacesLibrary> {
  if (placesLibraryPromise) return placesLibraryPromise;

  placesLibraryPromise = new Promise<google.maps.PlacesLibrary>(
    (resolve, reject) => {
      if (window.google?.maps?.places) {
        resolve(window.google.maps.places);
        return;
      }

      const callbackName = "__glassactGoogleMapsLoaded__";
      (window as unknown as Record<string, () => void>)[callbackName] = () => {
        resolve(google.maps.places);
      };

      const script = document.createElement("script");
      script.src = `https://maps.googleapis.com/maps/api/js?key=${encodeURIComponent(apiKey)}&libraries=places&loading=async&callback=${callbackName}`;
      script.async = true;
      script.onerror = () =>
        reject(new Error("Failed to load the Google Maps script."));
      document.head.appendChild(script);
    },
  );

  return placesLibraryPromise;
}
