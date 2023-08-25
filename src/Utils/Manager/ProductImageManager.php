<?php


namespace App\Utils\Manager;

use App\Entity\Product;
use App\Entity\ProductImage;
use App\Utils\File\ImageResizer;
use App\Utils\Filesystem\FilesystemWorker;
use Doctrine\ORM\EntityManagerInterface;
use Doctrine\Persistence\ObjectRepository;

class ProductImageManager extends AbstractBaseManager
{

    private FilesystemWorker $filesystemWorker;
    private string $uploadsTempDir;
    private ImageResizer $imageResizer;

    /**
     * @param EntityManagerInterface $entityManager
     * @param FilesystemWorker $filesystemWorker
     * @param ImageResizer $imageResizer
     * @param string $uploadsTempDir
     */
    public function __construct(
        EntityManagerInterface $entityManager,
        FilesystemWorker       $filesystemWorker,
        ImageResizer           $imageResizer,
        string                 $uploadsTempDir,
    )
    {
        parent::__construct($entityManager);
        $this->entityManager = $entityManager;
        $this->filesystemWorker = $filesystemWorker;
        $this->uploadsTempDir = $uploadsTempDir;
        $this->imageResizer = $imageResizer;
    }

    /**
     * @param string $productDir
     * @param string|null $tempImageFilename
     * @return ProductImage|null
     */
    public function resizeAndSaveImageForProduct(string $productDir, string $tempImageFilename = null): ?ProductImage
    {
        if (!$tempImageFilename) {
            return null;
        }

        $this->filesystemWorker->createFolder($productDir);

        $imageSmallParams = [
            'width' => 60,
            'height' => null,
            'newFolder' => $productDir,
            'newFileName' => sprintf('%s_%s.jpg', uniqid(), 'small'),
        ];
        $imageSmall = $this->imageResizer->resizeImageAndSave(
            $this->uploadsTempDir,
            $tempImageFilename,
            $imageSmallParams,
        );

        $imageMiddleParams = [
            'width' => 430,
            'height' => null,
            'newFolder' => $productDir,
            'newFileName' => sprintf('%s_%s.jpg', uniqid(), 'middle'),
        ];
        $imageMiddle = $this->imageResizer->resizeImageAndSave(
            $this->uploadsTempDir,
            $tempImageFilename,
            $imageMiddleParams,
        );

        $imageBigParams = [
            'width' => 800,
            'height' => null,
            'newFolder' => $productDir,
            'newFileName' => sprintf('%s_%s.jpg', uniqid(), 'big'),
        ];
        $imageBig = $this->imageResizer->resizeImageAndSave(
            $this->uploadsTempDir,
            $tempImageFilename,
            $imageBigParams,
        );

        $productImage = new ProductImage();
        $productImage->setFilenameSmall($imageSmall);
        $productImage->setFilenameMiddle($imageMiddle);
        $productImage->setFilenameBig($imageBig);

        return $productImage;
    }


    /**
     * @param ProductImage $productImage
     * @param string $productDir
     * @return void
     */
    public function removeImageFromProduct(
        ProductImage $productImage,
        string       $productDir,
    ): void
    {
        $product = $productImage->getProduct();

        $smallFilePath = $productDir . '/' . $productImage->getFilenameSmall();
        $this->filesystemWorker->remove($smallFilePath);

        $middleFilePath = $productDir . '/' . $productImage->getFilenameMiddle();
        $this->filesystemWorker->remove($middleFilePath);

        $bigFilePath = $productDir . '/' . $productImage->getFilenameBig();
        $this->filesystemWorker->remove($bigFilePath);

        $product->removeProductImage($productImage);
        $this->entityManager->flush();
    }

    /**
     * @return ObjectRepository
     */
    public function getRepository(): ObjectRepository
    {
        return $this->entityManager->getRepository(ProductImage::class);
    }
}
