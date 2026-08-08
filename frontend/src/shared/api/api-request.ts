export const apiRequest = async <T>(
  request: Promise<Response>,
  textError: string | undefined = 'Произошла ошибка запроса',
  callBackError?: () => void,
): Promise<T> => {
  const response = await request;

  if (!response.ok) {
    if (callBackError) {
      callBackError();
    }

    const errorText = await response.text();
    throw new Error(errorText || textError);
  }

  return response.json() as Promise<T>;
};
