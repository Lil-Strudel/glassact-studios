export interface FaqGroup {
  id: string;
  title: string;
  questions: { question: string; answer: string }[];
}

export const faqGroups: FaqGroup[] = [
  {
    id: "the-glass",
    title: "The glass",
    questions: [
      {
        question: "What does GlassAct Studios specialize in?",
        answer:
          "We make permanent color, custom glass inlays for the memorial industry. The glass is specially hardened and about as tough as the sandblasted granite.",
      },
      {
        question: "How special is your product?",
        answer:
          "Nobody in the industry has gone through all the steps and troubleshooting to produce a solid color glass inlay. The process and the inlays are patented.",
      },
      {
        question: "How old is your oldest inlay?",
        answer:
          "Our first glass inlay was installed in 2003. Even though the glass in that gravestone is still performing well, we have made many changes to strengthen our glass and the bond between glass and granite.",
      },
      {
        question: "How many different colors do you work with?",
        answer:
          "We have 40 different colors that will never fade. In special cases, we can order other colors.",
      },
      {
        question: "How do you get the color into the glass?",
        answer:
          "The color is not printed on the glass. We use two layers of colored fusible plate glass that are melted together, producing a ¼″ thick, solid colored glass.",
      },
      {
        question: "Where do you get your fusible glass?",
        answer:
          "We use the best glass manufacturers in the USA. We never use any Asian products.",
      },
      {
        question: "How hard is your glass?",
        answer:
          "Almost as hard as the granite. Our glass will scratch the granite, but granite can also scratch our glass.",
      },
      {
        question: "How tough is your glass?",
        answer:
          "The inlays are every bit as strong as sandblasted granite. Just like granite has its limitations, the glass does too. With our special fire-polishing and annealing method, the glass becomes 200 times stronger than regular glass. In our tests, the glass was as resilient as the sandblasted granite.",
      },
    ],
  },
  {
    id: "designs-and-ordering",
    title: "Designs and ordering",
    questions: [
      {
        question: "How expensive are the glass inlays?",
        answer:
          "We work in 4 different price groups, starting at $80 and up. We also sell the special glue, grout and other consumables in an install kit. We ask our installers to keystone pricing. Our product is a value-added product: you will be able to upcharge your client and therefore earn higher profits.",
      },
      {
        question: "Where will I get the designs from?",
        answer:
          "We have a catalog with hundreds of design options. This catalog is printed and also in digital format.",
      },
      {
        question: "I have designs I like, but I can’t find them in your catalog.",
        answer:
          "We will redesign existing sandblasting files into colored glass inlays.",
      },
      {
        question: "Do I have to hand fit my own stencils around the glass detail?",
        answer:
          "We create a 100% fitting stencil file. We test every design with a printout and we guarantee the fit.",
      },
      {
        question: "How long does it take to get the glass inlay after ordering it?",
        answer:
          "We usually ship the finished inlay within 2 weeks. After placing the order, a finishing date will be given.",
      },
      {
        question: "Are designs the only thing you do in glass?",
        answer:
          "We can create most fonts, if the font is at least 1″ tall and 1/8″–3/16″ wide.",
      },
    ],
  },
  {
    id: "care-and-warranty",
    title: "Care, durability and warranty",
    questions: [
      {
        question:
          "If a truck drives over my monument, do I have to worry that it will break my glass design?",
        answer:
          "Our glass is toughened to withstand everyday wear. Weather, trucks and lawn mowers are not a problem.",
      },
      {
        question: "What will happen to the glass if I use acidic stone cleaner?",
        answer:
          "The glass can be cleaned with any stone cleaners containing acids. The grout will get etched, but the glass remains untouched.",
      },
      {
        question: "Can the glass inlays ever be repaired if they get vandalized?",
        answer:
          "If the glass ever gets vandalized, replacements are made and easy to install. The glass cannot fall out of the void. It might get a crack or chip from a brutal impact, but the glass reacts like a car windshield: it stays together.",
      },
      {
        question: "What kind of warranty do you give?",
        answer:
          "GlassAct Studios produces custom designed and permanent colored glass to be inlaid in granite memorials. We guarantee that the glass is of the highest quality. If the installation of the glass is done using our specified materials and is finished according to the strict installation methods, we will guarantee the glass for colorfastness and craftsmanship for as long as GlassAct Studios, Inc. is in business. With the exception of acts of God, vandalism or abuse, or faulty installation, we will mail you a replacement of the glass, free of charge.",
      },
    ],
  },
];
