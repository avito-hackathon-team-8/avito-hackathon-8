import { leaderboardMock } from "../model/mock/leaderboard-mock";

export type TLeaderboardUser = {
  id: string;
  nickname: string;
  position: number;
};

export const getLeaderBoard = async (): Promise<TLeaderboardUser[]> => {
  return await Promise.resolve(leaderboardMock);
};
