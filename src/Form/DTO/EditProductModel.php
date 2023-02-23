<?php

namespace App\Form\DTO;

use App\Entity\Category;
use App\Entity\Product;
use http\Message;
use Symfony\Component\HttpFoundation\File\UploadedFile;
use Symfony\Component\Validator\Tests\Constraints as Asserts;

class EditProductModel
{

    public ?string $id;

    public string $title;

    public string $price;

    public ?UploadedFile $newImage;

    public int $quality;

    public string $description;

    public Category $category;

    public bool $isPublished;

    public bool $isDeleted;

    /**
     * @param Product|null $product
     * @return static
     */
    public static function makeFromProduct(?Product $product): self
    {
        $model = new self();

        if (!$product) {
            $model->id = null;

            return $model;
        }

        $model->id = $product->getId();
        $model->title = $product->getTitle();
        $model->price = $product->getPrice();
        $model->quality = $product->getQuality();
        $model->description = $product->getDescription();
        $model->isPublished = $product->isIsPublished();
        $model->isDeleted = $product->isIsDeleted();

        return $model;
    }
}
