// Guide bushing offsets, as Bushwhack applies them (see the Guide page).

/** Hole mode: the opening matches the drawing. */
export const holeOffset = (bushingIn: number, bitIn: number) => (bushingIn - bitIn) / 2;

/** Piece mode: the part that falls out matches the drawing. */
export const pieceOffset = (bushingIn: number, bitIn: number) => (bushingIn + bitIn) / 2;

export interface Combo {
	bushing: string;
	bit: string;
	bushingIn: number;
	bitIn: number;
}

const mm = 1 / 25.4;

/** Common bushing and bit pairings, for the reference tables. */
export const combos: Combo[] = [
	{ bushing: '5/16 in', bit: '1/8 in', bushingIn: 5 / 16, bitIn: 1 / 8 },
	{ bushing: '3/8 in', bit: '1/8 in', bushingIn: 3 / 8, bitIn: 1 / 8 },
	{ bushing: '7/16 in', bit: '1/4 in', bushingIn: 7 / 16, bitIn: 1 / 4 },
	{ bushing: '1/2 in', bit: '1/4 in', bushingIn: 1 / 2, bitIn: 1 / 4 },
	{ bushing: '5/8 in', bit: '1/4 in', bushingIn: 5 / 8, bitIn: 1 / 4 },
	{ bushing: '5/8 in', bit: '1/2 in', bushingIn: 5 / 8, bitIn: 1 / 2 },
	{ bushing: '3/4 in', bit: '1/2 in', bushingIn: 3 / 4, bitIn: 1 / 2 },
	{ bushing: '17 mm', bit: '6 mm', bushingIn: 17 * mm, bitIn: 6 * mm },
	{ bushing: '30 mm', bit: '12 mm', bushingIn: 30 * mm, bitIn: 12 * mm }
];
