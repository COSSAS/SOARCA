import React, { useState } from 'react';
import { X } from 'lucide-react';

interface BannerProps {
}

const Banner: React.FC<BannerProps> = () => {
  const [bannerOpen, setBannerOpen] = useState<boolean>(true);
  const query = new URLSearchParams(location.search);
  const template = query.get('template');

  const liteLink: string =
    template === 'laravel'
      ? 'https://github.com/cruip/laravel-tailwindcss-admin-dashboard-template'
      : 'https://github.com/cruip/tailwind-dashboard-template';

  return (
    <>
      {bannerOpen && (
        <div className="fixed bottom-0 right-0 w-full md:bottom-8 md:right-12 md:w-auto z-50">
          <div className="bg-gray-800 border border-transparent dark:border-gray-700/60 text-gray-50 text-sm p-3 md:rounded-sm shadow-lg flex justify-between">
            <div className="text-gray-500 inline-flex">
              <a
                className="font-medium hover:underline text-gray-50"
                href={liteLink}
                target="_blank"
                rel="noreferrer"
              >
                Download
                <span className="hidden sm:inline"> on GitHub</span>
              </a>
              <span className="italic px-1.5">or</span>
              <a
                className="font-medium hover:underline text-emerald-400"
                href="https://cruip.com/mosaic/"
                target="_blank"
                rel="noreferrer"
              >
                Check Premium Version
              </a>
            </div>
            <button
              className="text-gray-500 hover:text-gray-400 pl-2 ml-3 border-l border-gray-700/60"
              onClick={() => setBannerOpen(false)}
            >
              <span className="sr-only">Close</span>
              <X className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}
    </>
  );
};

export default Banner;
