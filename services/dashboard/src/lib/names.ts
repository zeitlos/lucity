const adjectives = [
  'brave', 'bright', 'bold', 'calm', 'clever',
  'cosmic', 'crisp', 'daring', 'eager', 'electric',
  'epic', 'fierce', 'flash', 'fresh', 'gentle',
  'golden', 'happy', 'humble', 'keen', 'lively',
  'lucky', 'magic', 'mighty', 'noble', 'nifty',
  'plucky', 'proud', 'quick', 'quiet', 'rapid',
  'rustic', 'sharp', 'silent', 'slick', 'smooth',
  'snappy', 'solar', 'sonic', 'spicy', 'steady',
  'stellar', 'stormy', 'super', 'swift', 'turbo',
  'vivid', 'warm', 'wild', 'witty', 'zen',
];

const animals = [
  'badger', 'bear', 'bison', 'bobcat', 'cobra',
  'crane', 'crow', 'dingo', 'dolphin', 'eagle',
  'falcon', 'ferret', 'fox', 'gecko', 'hawk',
  'heron', 'husky', 'ibis', 'jaguar', 'koala',
  'lemur', 'lion', 'lynx', 'manta', 'moose',
  'narwhal', 'otter', 'owl', 'panda', 'parrot',
  'pelican', 'phoenix', 'pike', 'puma', 'quail',
  'raven', 'robin', 'salmon', 'shark', 'sloth',
  'sparrow', 'squid', 'stork', 'tiger', 'toucan',
  'viper', 'walrus', 'wolf', 'wren', 'yak',
];

function pick<T>(arr: T[]): T {
  return arr[Math.floor(Math.random() * arr.length)];
}

/**
 * Generate a fun human-readable name like "cosmic-panda" or "swift-falcon".
 */
export function generateName(): string {
  return `${pick(adjectives)}-${pick(animals)}`;
}
