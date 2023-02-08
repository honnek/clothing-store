<?php


namespace App\Utils\Filesystem;

use Symfony\Component\Filesystem\Filesystem;

class FilesystemWorker
{

    private Filesystem $filesystem;

    public function __construct(Filesystem $filesystem)
    {
        $this->filesystem = $filesystem;
    }

    /**
     * @param string $folder
     * @return void
     */
    public function createFolder(string $folder): void
    {
        if (!$this->filesystem->exists($folder)) {
            $this->filesystem->mkdir($folder);
        }
    }

    /**
     * @param string $item
     * @return void
     */
    public function remove(string $item)
    {
        if ($this->filesystem->exists($item)) {
            $this->filesystem->remove($item);
        }
    }

}

