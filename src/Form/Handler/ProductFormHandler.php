<?php

namespace App\Form\Handler;

use App\Entity\Product;
use App\Utils\File\FileSaver;
use App\Utils\Manager\ProductManager;
use Doctrine\ORM\EntityManagerInterface;
use Symfony\Component\Form\Form;
use Symfony\Component\Form\FormInterface;

class ProductFormHandler
{

    private ProductManager $productManager;
    private FileSaver $fileSaver;

    public function __construct(ProductManager $productManager, FileSaver $fileSaver)
    {
        $this->productManager = $productManager;
        $this->fileSaver = $fileSaver;
    }

    /**
     * @TODO Add a new image with diff size
     * 1. сохранение изменений продукта
     * 2. Сохранение загруженного файла в temp folder
     *
     * 3. Провидим работу с продуктом (добавляем изображение в него)
     * 3.1 получаем путь к папке с картинками
     * 3.2 изменение рамера и сохранение картинки в папки (BIG, MEDIUM, small)
     * 3.3 созданем ProductImage и возвращем его продукту
     * 3.4 Сохраняем продукт вместе с новым productImage
     *
     * @param Product $product
     * @param FormInterface $form
     * @return Product
     */
    public function processEditForm(Product $product, FormInterface $form): Product
    {
        /** сохранение изменений продукта */
        $this->productManager->save($product);

        $newImageFile = $form->get('newImage')->getData();
        $tempImageFile = $newImageFile ? $this->fileSaver->saveUploadedFileIntoTemp($newImageFile) : null;

        $this->productManager->updateProductImages($product, $tempImageFile);

        $this->productManager->save($product);
dd($product);
        return $product;
    }
}
