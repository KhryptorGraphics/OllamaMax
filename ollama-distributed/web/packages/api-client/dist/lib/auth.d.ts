export interface LoginReq {
    username: string;
    password: string;
}
export interface LoginRes {
    token: string;
    user: any;
    session_id?: string;
    expires_at?: string;
}
export interface RegisterReq {
    username: string;
    email: string;
    password: string;
}
export interface ForgotPasswordReq {
    email: string;
}
export interface ResetPasswordReq {
    token: string;
    password: string;
}
export interface VerifyEmailReq {
    token: string;
}
export declare const AuthAPI: {
    login: (payload: LoginReq) => Promise<LoginRes>;
    register: (payload: RegisterReq) => Promise<any>;
    logout: () => Promise<{
        message: string;
    }>;
    profile: () => Promise<{
        user: any;
    }>;
    forgotPassword: (payload: ForgotPasswordReq) => Promise<{
        message: string;
    }>;
    resetPassword: (payload: ResetPasswordReq) => Promise<{
        message: string;
    }>;
    verifyEmail: (payload: VerifyEmailReq) => Promise<{
        message: string;
    }>;
};
