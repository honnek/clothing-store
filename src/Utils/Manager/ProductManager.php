<?php

namespace App\Utils\Manager;


use App\Entity\Product;
use Doctrine\ORM\EntityManagerInterface;
use Doctrine\Persistence\ObjectRepository;

class ProductManager extends AbstractBaseManager
{

    private string $productImagesDir;

    private ProductImageManager $productImageManager;

    public function __construct(
        EntityManagerInterface $entityManager,
        ProductImageManager    $productImageManager,
        string                 $productImagesDir,
    )
    {
        parent::__construct($entityManager);
        $this->productImagesDir = $productImagesDir;
        $this->productImageManager = $productImageManager;
    }

    /**
     * @return ObjectRepository
     */
    public function getRepository(): ObjectRepository
    {
        return $this->entityManager->getRepository(Product::class);
    }

    /**
     * @param object $product
     * @return void
     */
    public function remove(object $product): void
    {
        $product->setIsDeleted(true);
        $this->save($product);
    }

    /**
     * @param Product $product
     * @return string
     */
    public function getProductImagesDir(Product $product): string
    {
        return sprintf('%s/%s', $this->productImagesDir, $product->getId());
    }

    public function updateProductImages(Product $product, string $tempImageFilename = null): Product
    {
        if (!$tempImageFilename) {
            return $product;
        }

        $productDir = $this->getProductImagesDir($product);
        $productImage = $this->productImageManager->resizeAndSaveImageForProduct($productDir, $tempImageFilename);
        $productImage->setProduct($product);
        $product->addProductImage($productImage);

        return $product;
    }
}
