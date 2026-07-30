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

      if (!apiKey) {
        reject(
          new Error(
            "VITE_GOOGLE_MAPS_API_KEY is not set, so address search cannot load.",
          ),
        );
        return;
      }

      const callbackName = "__glassactGoogleMapsLoaded__";
      (window as unknown as Record<string, () => void>)[callbackName] = () => {
        resolve(google.maps.places);
      };

      // A rejected key still serves a valid script, so onerror never fires and
      // the callback never runs — without this the promise hangs forever and
      // the field just sits there returning no suggestions.
      (window as unknown as Record<string, () => void>).gm_authFailure = () =>
        reject(
          new Error(
            "Google Maps rejected the API key. Check its HTTP referrer restrictions and that the Places API is enabled.",
          ),
        );

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
