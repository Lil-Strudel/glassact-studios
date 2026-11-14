import type { ImageMetadata } from "astro";
import animalsImage from "../assets/images/gallery/11-Zoo6x6-300_600.jpg";
import outdoorsImage from "../assets/images/gallery/15-jeep6x6-300_600.jpg";
import childrenImage from "../assets/images/gallery/1-Ang6x6-300_600.jpg";
import flowersImage from "../assets/images/gallery/4-Daf6x6-300_600.jpg";
import faithImage from "../assets/images/gallery/7-HT6x6-300_600.jpg";
import ornamentsImage from "../assets/images/gallery/3-Celt6x6-300_600.jpg";

export interface CatalogCategory {
  title: string;
  description: string;
  page?: number;
  image: ImageMetadata;
  imageAlt: string;
}

export const catalogCategories: CatalogCategory[] = [
  {
    title: "Animals",
    description: "Birds, dogs, cats, butterflies and more",
    page: 8,
    image: animalsImage,
    imageAlt: "Glass inlay of an elephant, giraffe and raccoon set in granite",
  },
  {
    title: "Outdoors",
    description: "Fishing, cowboys, classic cars and more",
    page: 26,
    image: outdoorsImage,
    imageAlt: "Glass inlay of a jeep set in granite",
  },
  {
    title: "Children",
    description: "Cute, wholesome, tender and sweet",
    page: 50,
    image: childrenImage,
    imageAlt: "Glass inlay of a sleeping baby angel on a cloud",
  },
  {
    title: "Flowers",
    description: "Roses, daisies, tulips and more",
    page: 66,
    image: flowersImage,
    imageAlt: "Glass inlay of yellow daffodils set in granite",
  },
  {
    title: "Faith",
    description: "Temples, doves and other religious symbols",
    page: 128,
    image: faithImage,
    imageAlt: "Glass inlay of a cross on a red heart between two flowers",
  },
  {
    title: "And more",
    description: "Ornaments, professions and more",
    image: ornamentsImage,
    imageAlt: "Glass inlay of a Celtic knot set in granite",
  },
];

export function catalogHref(category: CatalogCategory) {
  return category.page ? `/catalog?page=${category.page}` : "/catalog";
}
