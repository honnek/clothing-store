<?php


namespace App\Utils\File;

use App\Utils\Filesystem\FilesystemWorker;
use http\Exception\RuntimeException;
use Symfony\Component\Filesystem\Filesystem;
use Symfony\Component\HttpFoundation\File\Exception\FileException;
use Symfony\Component\HttpFoundation\File\UploadedFile;
use Symfony\Component\String\Slugger\SluggerInterface;

class FileSaver
{
    private SluggerInterface $slugger;
    private string $uploadsTempDir;
    private FilesystemWorker $filesystemWorker;

    public function __construct(
        SluggerInterface $slugger,
        FilesystemWorker $filesystemWorker,
        string           $uploadsTempDir,
    )
    {
        $this->slugger = $slugger;
        $this->uploadsTempDir = $uploadsTempDir;
        $this->filesystemWorker = $filesystemWorker;
    }

    /**
     * @param UploadedFile $uploadedFile
     * @return string|null
     */
    public function saveUploadedFileIntoTemp(UploadedFile $uploadedFile): ?string
    {
        $originalFileName = pathinfo($uploadedFile->getClientOriginalName(), PATHINFO_FILENAME);
        $safeFileName = $this->slugger->slug($originalFileName);
        $fileName = sprintf('%s-%s.%s', $safeFileName, uniqid(), $uploadedFile->guessExtension());

        $this->filesystemWorker->createFolder($this->uploadsTempDir);

        try {
            $uploadedFile->move($this->uploadsTempDir, $fileName);
        } catch (FileException $exception) {
            throw new RuntimeException('Не удается поместить загруженный файл: '.
                $fileName .
                $exception->getMessage()
            );
        }

        return $fileName;
    }


}