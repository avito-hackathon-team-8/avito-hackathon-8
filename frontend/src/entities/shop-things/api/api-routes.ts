export const API_ROUTE_SHOP = {
  items: '/v1/shop',

  purchase: (itemId: string) => `/v1/shop/${itemId}/purchase`,
};
