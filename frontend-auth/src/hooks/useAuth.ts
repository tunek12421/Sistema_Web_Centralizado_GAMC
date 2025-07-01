import { useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { selectAuth, verifyToken } from '../store/authSlice';

export const useAuth = () => {
  const dispatch = useDispatch() as any;
  const auth = useSelector(selectAuth);

  useEffect(() => {
    const token = localStorage.getItem('accessToken');
    if (token && !auth.user) {
      dispatch(verifyToken());
    }
  }, [dispatch, auth.user]);

  return {
    user: auth.user,
    isAuthenticated: auth.isAuthenticated,
    loading: auth.loading,
    error: auth.error,
  };
};
