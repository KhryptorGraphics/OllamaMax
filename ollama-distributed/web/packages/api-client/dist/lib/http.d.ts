export interface RequestOptions extends RequestInit {
    baseUrl?: string;
}
export declare function http<T>(path: string, { baseUrl, headers, ...init }?: RequestOptions): Promise<T>;
