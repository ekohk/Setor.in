/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // Allow images from S3/MinIO (later, when upload sprint lands)
  images: {
    remotePatterns: [
      { protocol: 'http', hostname: 'localhost' },
    ],
  },
};

export default nextConfig;
